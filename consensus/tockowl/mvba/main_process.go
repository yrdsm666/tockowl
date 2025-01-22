package mvba

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"
)

const minProposalSize = 5

// MainProcess is the main process of smvba instances
func MainProcess(ctx context.Context, p *party.HonestParty, epoch []byte, value []byte, validation []byte, Q func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error) []byte {

	haltChannel := make(chan []byte, 1024) //control all round
	hashVerifyMap := sync.Map{}
	sigVerifyMap := sync.Map{}
	bd := false // 𝑏𝑑 indicates whether 𝑝𝑖 has broadcast data

	parentQc1 := &protobuf.Lock{}
	parentQc2 := &protobuf.Lock{}
	preRoundSeed := 0

	for r := uint32(0); ; r++ {
		spbCtx, spbCancel := context.WithCancel(ctx) //son of ctx
		wg := sync.WaitGroup{}
		wg.Add(int(p.N + 1)) // n TPBReceiver and 1 TPBSender instances

		V := sync.Map{}
		Q1 := sync.Map{}
		Q2 := sync.Map{}
		Q3 := sync.Map{}

		doneFlagChannel := make(chan bool, 1)
		seedChannel := make(chan int, 1)  // for main progress
		bestChannel := make(chan bool, 1) // for main progress

		// Initialize three-phase broadcast instances
		var buf bytes.Buffer
		buf.Write(epoch)
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
				var safeProposal func(uint32, uint32, int) bool
				if r == 0 {
					validator = nil
					safeProposal = nil
				} else {
					validator = Q
					safeProposal = SafeProposal
				}

				// Receive broadcast values from other replicas
				_ = tpbReceiver(spbCtx, p, j, IDrj[j], &V, &Q1, &Q2, validator, safeProposal, parentQc2.Replica, preRoundSeed, &hashVerifyMap, &sigVerifyMap)
				wg.Done()
			}(i)
		}

		// Run this party's TPB instance
		go func() {
			var tpbValue []byte
			var tpbSig []byte
			var ok bool
			// A replica broadcasts its own value
			if !p.FastTrack {
				// In TockOwl, all replicas start the three-phase broadcast at the same time
				tpbValue, tpbSig, ok = tpbSender(spbCtx, p, IDrj[p.PID], value, validation, parentQc1)
			} else {
				// In TockOwl+, leader first start the three-phase broadcast
				nowEpoch := utils.BytesToUint32(epoch)
				if p.PID == p.GetLeaderByEpoch(nowEpoch) {
					// The leader immediately broadcasts its value
					bd = true
					tpbValue, tpbSig, ok = tpbSender(spbCtx, p, IDrj[p.PID], value, validation, parentQc1)
				} else {
					// Other replicas broadcast their values ​​only after the timer expires.
					time.Sleep(time.Duration(p.TimeOut) * time.Millisecond)
					bd = true
					tpbValue, tpbSig, ok = tpbSender(spbCtx, p, IDrj[p.PID], value, validation, parentQc1)
				}
			}

			if ok {
				finishMessage := core.Encapsulation("Finish", IDr, p.PID, &protobuf.Finish{
					Value: tpbValue,
					Sig:   tpbSig,
				})
				p.Broadcast(finishMessage)
			}
			wg.Done()
		}()

		// Run Message Handlers
		go messageHandler(ctx, p, epoch, IDr, IDrj, &bd,
			&V, &Q1, &Q2, &Q3,
			doneFlagChannel, bestChannel, seedChannel, haltChannel, r)

		// doneFlag -> common coin
		go election(ctx, p, IDr, doneFlagChannel)

		// seedChannel -> best
		select {
		case result := <-haltChannel:
			spbCancel()
			return result
		case seed := <-seedChannel:
			// Stop all three-phase broadcast instances
			spbCancel()
			preRoundSeed = seed
			wg.Wait()

			bestVOwner, value0, _ := BestElement(&V, seed, p, utils.BytesToUint32(epoch))
			bestQ1Owner, value1, sig1 := BestElement(&Q1, seed, p, utils.BytesToUint32(epoch))
			bestQ2Owner, value2, sig2 := BestElement(&Q2, seed, p, utils.BytesToUint32(epoch))
			bestQ3Owner, value3, sig3 := BestElement(&Q3, seed, p, utils.BytesToUint32(epoch))

			if p.PID == 1 {
				p.Log.Tracef("bv = %d, bq1 = %d, bq2 = %d, bq3 = %d", bestVOwner, bestQ1Owner, bestQ2Owner, bestQ3Owner)
			}

			bestMessage := core.Encapsulation("Best", IDr, p.PID, &protobuf.Best{
				MaxValue:      value0,
				MaxValueOwner: bestVOwner,
				MaxQ1Sig:      sig1,
				MaxQ1Value:    value1,
				MaxQ1Owner:    bestQ1Owner,
				MaxQ2Sig:      sig2,
				MaxQ2Value:    value2,
				MaxQ2Owner:    bestQ2Owner,
				MaxQ3Sig:      sig3,
				MaxQ3Value:    value3,
				MaxQ3Owner:    bestQ3Owner,
			})
			p.Broadcast(bestMessage)

			// Result
			select {
			case result := <-haltChannel:
				return result
			case <-bestChannel:
				replica0, _, _ := BestElement(&V, seed, p, utils.BytesToUint32(epoch))
				replica1, value1, sig1 := BestElement(&Q1, seed, p, utils.BytesToUint32(epoch))
				replica2, _, _ := BestElement(&Q2, seed, p, utils.BytesToUint32(epoch))
				replica3, value3, sig3 := BestElement(&Q3, seed, p, utils.BytesToUint32(epoch))
				priority0 := CalculateReplicaPriority(replica0, seed)
				priority1 := CalculateReplicaPriority(replica1, seed)
				priority2 := CalculateReplicaPriority(replica2, seed)
				priority3 := CalculateReplicaPriority(replica3, seed)

				// Next round value
				value := value1
				validation = sig1

				// Commit rules
				if priority0 == priority1 && priority1 == priority2 && priority2 == priority3 && replica0 == replica1 && replica1 == replica2 && replica2 == replica3 {
					// if priority0 == priority3 && replica0 == replica3 {
					haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
						Replica: replica3,
						Value:   value3,
						Sig:     sig3,
					})
					p.Broadcast(haltMessage)
					return value
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

		coinShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, coinName) //sign("Done"||epoch||r||coin share)
		doneMessage := core.Encapsulation("Done", IDr, p.PID, &protobuf.Done{
			CoinShare: coinShare,
		})

		p.Broadcast(doneMessage)

	}
}

func SafeProposal(sender uint32, parentQc2Owner uint32, preSeed int) bool {
	priority0 := CalculateReplicaPriority(sender, preSeed)
	priority1 := CalculateReplicaPriority(parentQc2Owner, preSeed)
	return priority0 >= priority1
}
