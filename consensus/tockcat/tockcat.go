package tockcat

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
	haltChannel := make(chan []byte, 1024) // control all round

	hashVerifyMap := sync.Map{}
	sigVerifyMap := sync.Map{}

	for r := uint32(0); ; r++ {
		spbCtx, spbCancel := context.WithCancel(ctx) //son of ctx
		wg := sync.WaitGroup{}
		wg.Add(int(p.N + 1)) // n SPBReceiver and 1 SPBSender instances

		doneFlagChannel := make(chan bool, 1)
		leaderChannel := make(chan uint32, 1)  // for main progress
		voteFlagChannel := make(chan byte, 1)  // for main progress
		voteYesChannel := make(chan []byte, 2) // for main progress
		voteNoChannel := make(chan []byte, 1)  // for main progress

		V := sync.Map{}
		Q1 := sync.Map{}
		Q2 := sync.Map{}

		// Initialize two-phase broadcast instances
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

				// Receive broadcast values from other replicas
				value, sig, ok := spbReceiver(spbCtx, p, j, IDrj[j], validator, &hashVerifyMap, &sigVerifyMap)
				if ok { //save Lock
					V.Store(j, &protobuf.Lock{
						Value: value,
					})
					Q1.Store(j, &protobuf.Lock{
						Value: value,
						Sig:   sig,
					})
				}
				wg.Done()
			}(i)
		}

		// Run this party's SPB instance
		go func() {
			// A replica broadcasts its own value
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

		// Run Message Handlers
		go messageHandler(ctx, p, IDr, IDrj,
			&Q2,
			doneFlagChannel,
			voteFlagChannel, voteYesChannel, voteNoChannel,
			leaderChannel, haltChannel, r)

		// doneFlag -> common coin
		go election(ctx, p, IDr, doneFlagChannel)

		// leaderChannel -> shortcut -> prevote||vote||viewchange
		select {
		case result := <-haltChannel:
			spbCancel()
			return result
		case l := <-leaderChannel:
			spbCancel()
			wg.Wait()

			// shortcut
			value1, ok1 := Q2.Load(l)
			if ok1 && p.Shortcut {
				finish := value1.(*protobuf.Finish)
				haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
					Value: finish.Value,
					Sig:   finish.Sig,
				})
				p.Broadcast(haltMessage)
				if p.PID == 1 {
					p.Log.Tracef("[p=%d,r=%d]: Shortcut: Commit a value: Q2 set has the leader %d value", p.PID, r, l)
				}
				return finish.Value
			}

			// bestExchange
			go bestExchange(p, IDr, l, &Q1, &Q2)

			// Result
			select {
			case result := <-haltChannel:
				return result
			case flag := <-voteFlagChannel:
				value1, ok1 := Q2.Load(l)
				if ok1 {
					finish := value1.(*protobuf.Finish)
					haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
						Value: finish.Value,
						Sig:   finish.Sig,
					})
					p.Broadcast(haltMessage)
					if p.PID == 1 {
						p.Log.Tracef("[p=%d,r=%d]: After Best: Commit a value for leader %d value", p.PID, r, l)
					}
					return finish.Value
				}

				// Entering a new iteration
				if flag == 0 {
					value = <-voteYesChannel
					validation = <-voteYesChannel
					if p.PID == 1 {
						p.Log.Tracef("[p=%d,r=%d]: Inter a new round: Receive a vaild Best message with leader value", p.PID, r)
					}
				} else {
					sig := <-voteNoChannel
					validation = append(validation, sig...)
					if p.PID == 1 {
						p.Log.Tracef("[p=%d,r=%d]: Inter a new round: All Best message without leader value", p.PID, r)
					}
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

func bestExchange(p *party.HonestParty, IDr []byte, l uint32, Q1 *sync.Map, Q2 *sync.Map) {
	var leaderSig2 []byte
	value2, ok2 := Q2.Load(l)
	if ok2 && !p.IsByzantineNode(int64(p.PID)) { // Byzantine replicas do not broadcast valid BestMsg messages
		commit := value2.(*protobuf.Finish)
		leaderSig2 = commit.Sig
	}

	value1, ok1 := Q1.Load(l)
	if ok1 && !p.IsByzantineNode(int64(p.PID)) { // Byzantine replicas do not broadcast valid BestMsg messages
		lock := value1.(*protobuf.Lock)
		bestMessage := core.Encapsulation("TockCatBest", IDr, p.PID, &protobuf.TockCatBest{
			Sig1:   lock.Sig,
			Sig2:   leaderSig2,
			Value:  lock.Value,
			Leader: utils.Uint32ToBytes(l),
		})
		p.Broadcast(bestMessage)
	} else {
		var buf bytes.Buffer
		buf.Write([]byte("BAD LEADER"))
		buf.Write(IDr)
		sm := buf.Bytes()
		sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign(false||ID||r)
		bestMessage := core.Encapsulation("TockCatBest", IDr, p.PID, &protobuf.TockCatBest{
			Sigshare: sigShare,
			Leader:   utils.Uint32ToBytes(l),
		})
		p.Broadcast(bestMessage)
	}
}
