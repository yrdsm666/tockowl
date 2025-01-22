package mvba

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"

	"golang.org/x/crypto/sha3"
)

func messageHandler(ctx context.Context, p *party.HonestParty, epoch []byte, IDr []byte, IDrj [][]byte,
	bd *bool,
	V *sync.Map, Q1 *sync.Map, Q2 *sync.Map, Q3 *sync.Map,
	doneFlagChannel chan bool, bestChannel chan bool, seedChannel chan int, haltChannel chan []byte, r uint32) {

	// FinishMessage Handler
	go func() {
		Q3Length := 0
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Finish", IDr):
				payload := core.Decapsulation("Finish", m).(*protobuf.Finish)
				h := sha3.Sum512(payload.Value)
				var buf bytes.Buffer
				buf.Write([]byte("Echo"))
				buf.Write(IDrj[m.Sender])
				buf.WriteByte(3)
				buf.Write(h[:])
				sm := buf.Bytes()
				err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||j||2||h)
				if err == nil {
					Q3.Store(m.Sender, payload)

					// Output from fast track
					if p.FastTrack {
						nowEpoch := utils.BytesToUint32(epoch)
						if m.Sender == p.GetLeaderByEpoch(nowEpoch) && *bd {
							// Leader's proposal cannot be empty if lightLoad is open
							if !p.LightLoad || len(payload.Value) > minProposalSize {
								haltMessage := core.Encapsulation("Halt", IDr, p.PID, &protobuf.Halt{
									Replica: m.Sender,
									Value:   payload.Value,
									Sig:     payload.Sig,
								})
								p.Broadcast(haltMessage)
								haltChannel <- payload.Value
							}
						}
					}

					Q3Length++
					if Q3Length == int(p.N-p.F) {
						doneFlagChannel <- true
					}
				} else {
					p.Log.Warn("Receive invalid Finish message.")
				}
			}
		}
	}()

	// DoneMessage Handler
	go func() {
		var buf bytes.Buffer
		buf.Write([]byte("Done"))
		buf.Write(IDr)
		coins := [][]byte{}
		ids := []int64{}
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Done", IDr):
				payload := core.Decapsulation("Done", m).(*protobuf.Done)
				coins = append(coins, payload.CoinShare)
				ids = append(ids, int64(m.Sender)+1)

				if len(coins) >= int(p.N-p.F) {
					coin, err := crypto.CombineSignatures(coins, ids)
					if err != nil {
						fmt.Println("error:", err)
					}
					seed := utils.BytesToInt(coin) // seed of round r
					seedChannel <- seed            // for main process
					p.Log.Trace("Seed: ", seed)
					return
				}
			}
		}
	}()

	// HaltMessage Handler
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Halt", IDr):
				payload := core.Decapsulation("Halt", m).(*protobuf.Halt)
				if payload.Replica >= p.N {
					p.Log.Warn("Receive invaild Halt message.")
					continue
				}

				h := sha3.Sum512(payload.Value)
				var buf bytes.Buffer
				buf.Write([]byte("Echo"))
				buf.Write(IDrj[payload.Replica])
				buf.WriteByte(3)
				buf.Write(h[:])
				sm := buf.Bytes()
				err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||l||2||h)
				if err == nil {
					haltChannel <- payload.Value
					return
				} else {
					p.Log.Warn("Verify Halt message failed.")
				}
			}
		}

	}()

	// BestMessage Handler
	go func() {
		bestMessageNumber := 0
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Best", IDr):
				payload := core.Decapsulation("Best", m).(*protobuf.Best)
				bestMessageNumber += 1

				V.Store(payload.MaxValueOwner, &protobuf.Lock{
					Value: payload.MaxValue,
				})
				Q1.Store(payload.MaxQ1Owner, &protobuf.Lock{
					Value: payload.MaxQ1Value,
					Sig:   payload.MaxQ1Sig,
				})
				Q2.Store(payload.MaxQ2Owner, &protobuf.Lock{
					Value: payload.MaxQ2Value,
					Sig:   payload.MaxQ2Sig,
				})
				Q3.Store(payload.MaxQ3Owner, &protobuf.Lock{
					Value: payload.MaxQ3Value,
					Sig:   payload.MaxQ3Sig,
				})
				if bestMessageNumber >= int(p.N-p.F) {
					bestChannel <- true
				}

			}
		}
	}()

}
