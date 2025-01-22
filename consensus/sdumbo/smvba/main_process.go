package smvba

import (
	"bytes"
	"context"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"
)

// MainProcess is the main process of smvba instances
func MainProcess(ctx context.Context, p *party.HonestParty, ID []byte, value []byte, validation []byte, Q func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error) []byte {
	haltChannel := make(chan []byte, 1024)       //control all round
	sonCtx, sonCancel := context.WithCancel(ctx) //son of ctx

	hashVerifyMap := sync.Map{}
	sigVerifyMap := sync.Map{}

	for r := uint32(0); ; r++ {
		spbCtx, spbCancel := context.WithCancel(ctx) //son of ctx
		wg := sync.WaitGroup{}
		wg.Add(int(p.N + 1)) //n SPBReceiver and 1 SPBSender instances

		Lr := sync.Map{} //Lock Set
		Fr := sync.Map{} //Finish Set
		doneFlagChannel := make(chan bool, 1)
		leaderChannel := make(chan uint32, 1)    //for main progress
		preVoteFlagChannel := make(chan bool, 1) //Yes or No
		preVoteYesChannel := make(chan []byte, 3)
		preVoteNoChannel := make(chan []byte, 2)
		voteFlagChannel := make(chan byte, 1)
		voteYesChannel := make(chan []byte, 2)
		voteNoChannel := make(chan []byte, 1)
		voteOtherChannel := make(chan []byte, 1)

		//Initialize SPB instances
		var buf bytes.Buffer
		buf.Write(ID)
		buf.Write(utils.Uint32ToBytes(r))
		IDr := buf.Bytes()

		IDrj := make([][]byte, 0, p.N)
		for j := uint32(0); j < p.N; j++ {
			var buf bytes.Buffer
			buf.Write(IDr)
			buf.Write(utils.Uint32ToBytes(j))
			IDrj = append(IDrj, buf.Bytes())
		}

		for i := uint32(0); i < p.N; i++ {
			go func(j uint32) {
				var validator func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error
				if r == 0 {
					validator = Q
				} else {
					validator = nil // TODO
				}
				value, sig, ok := spbReceiver(spbCtx, p, j, IDrj[j], validator, &hashVerifyMap, &sigVerifyMap)
				if ok { //save Lock
					Lr.Store(j, &protobuf.Lock{
						Value: value,
						Sig:   sig,
					})
				}
				wg.Done()
			}(i)
		}

		//Run this party's SPB instance
		go func() {
			value, sig, ok := spbSender(spbCtx, p, IDrj[p.PID], value, validation)
			if ok {
				finishMessage := core.Encapsulation("Finish", IDr, p.PID, &protobuf.Finish{
					Value: value,
					Sig:   sig,
				})
				p.Broadcast(finishMessage)
			}

			wg.Done()
		}()

		//Run Message Handlers
		go messageHandler(sonCtx, p, IDr, IDrj, &Fr,
			doneFlagChannel,
			preVoteFlagChannel, preVoteYesChannel, preVoteNoChannel,
			voteFlagChannel, voteYesChannel, voteNoChannel, voteOtherChannel,
			leaderChannel, haltChannel, r)

		//doneFlag -> common coin
		go election(sonCtx, p, IDr, doneFlagChannel)

		//leaderChannel -> shortcut -> prevote||vote||viewchange
		select {
		case result := <-haltChannel:
			spbCancel()
			sonCancel()
			return result
		case l := <-leaderChannel:
			spbCancel()
			wg.Wait()

			//short-cut
			value1, ok1 := Fr.Load(l)
			if ok1 && p.Shortcut {
				finish := value1.(*protobuf.Finish)
				haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
					Value: finish.Value,
					Sig:   finish.Sig,
				})
				if p.PID == 1 {
					p.Log.Tracef("[p=%d,r=%d]: Shortcut: Commit a value with leader %d value", p.PID, r, l)
				}
				p.Broadcast(haltMessage)
				sonCancel()
				return finish.Value
			}

			//preVote
			go preVote(sonCtx, p, IDr, l, &Lr)

			//vote
			go vote(sonCtx, p, IDr, l, preVoteFlagChannel, preVoteYesChannel, preVoteNoChannel, r)

			//result
			select {
			case result := <-haltChannel:
				sonCancel()
				return result
			case flag := <-voteFlagChannel:

				if flag == 0 { //Yes
					value := <-voteYesChannel
					sig := <-voteYesChannel
					haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
						Value: value,
						Sig:   sig,
					})
					p.Broadcast(haltMessage)
					sonCancel()
					return value
				} else if flag == 1 { //No
					sig := <-voteNoChannel
					validation = append(validation, sig...)
				} else {
					//overwrite
					value = <-voteOtherChannel
					validation = <-voteOtherChannel
				}
			}
		}
	}
}

func election(ctx context.Context, p *party.HonestParty, IDr []byte, doneFlageChannel chan bool) {
	select {
	case <-ctx.Done():
		return
	case <-doneFlageChannel:
		var buf bytes.Buffer
		buf.Write([]byte("Done"))
		buf.Write(IDr)
		coinName := buf.Bytes()

		coinShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, coinName)
		doneMessage := core.Encapsulation("Done", IDr, p.PID, &protobuf.Done{
			CoinShare: coinShare,
		})

		p.Broadcast(doneMessage)

	}
}

func preVote(ctx context.Context, p *party.HonestParty, IDr []byte, l uint32, Lr *sync.Map) {
	value2, ok2 := Lr.Load(l)
	if ok2 && !p.IsByzantineNode(int64(p.PID)) { // Byzantine replicas do not broadcast valid PreVote messages
		lock := value2.(*protobuf.Lock)
		preVoteMessage := core.Encapsulation("PreVote", IDr, p.PID, &protobuf.PreVote{
			Vote:  true,
			Value: lock.Value,
			Sig:   lock.Sig,
		})
		p.Broadcast(preVoteMessage)
	} else {
		var buf bytes.Buffer
		buf.WriteByte(byte(0)) //false
		buf.Write(IDr)
		sm := buf.Bytes()
		sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm)
		preVoteMessage := core.Encapsulation("PreVote", IDr, p.PID, &protobuf.PreVote{
			Vote:  false,
			Value: nil,
			Sig:   sigShare,
		})
		p.Broadcast(preVoteMessage)
	}
}

func vote(ctx context.Context, p *party.HonestParty, IDr []byte, l uint32, preVoteFlagChannel chan bool, preVoteYesChannel chan []byte, preVoteNoChannel chan []byte, r uint32) {
	select {
	case <-ctx.Done():
		return
	case VoteFlag := <-preVoteFlagChannel:
		if VoteFlag { //have received a valid Yes PreVote
			value := <-preVoteYesChannel
			sig := <-preVoteYesChannel
			sigShare := <-preVoteYesChannel
			voteMessage := core.Encapsulation("Vote", IDr, p.PID, &protobuf.Vote{
				Vote:     true,
				Value:    value,
				Sig:      sig,
				Sigshare: sigShare,
			})
			p.Broadcast(voteMessage)
		} else { //have received 2f+1 valid No PreVote
			sig := <-preVoteNoChannel
			sigShare := <-preVoteNoChannel
			voteMessage := core.Encapsulation("Vote", IDr, p.PID, &protobuf.Vote{
				Vote:     false,
				Value:    nil,
				Sig:      sig,
				Sigshare: sigShare,
			})
			p.Broadcast(voteMessage)
		}
	}
}
