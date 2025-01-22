package tockcat

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

func messageHandler(ctx context.Context, p *party.HonestParty, IDr []byte, IDrj [][]byte,
	Q2 *sync.Map,
	doneFlagChannel chan bool,
	voteFlagChannel chan byte, voteYesChannel chan []byte, voteNoChannel chan []byte,
	leaderChannel chan uint32, haltChannel chan []byte, r uint32) {

	// FinishMessage Handler
	go func() {
		FrLength := 0
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
				buf.WriteByte(2)
				buf.Write(h[:])
				sm := buf.Bytes()
				err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||j||2||h)
				if err == nil {
					Q2.Store(m.Sender, payload)
					FrLength++
					if FrLength == int(p.N-p.F) {
						doneFlagChannel <- true
					}
				}
			}
		}
	}()

	thisRoundLeader := make(chan uint32, 1)

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

				if len(coins) == int(p.F+1) {
					doneFlagChannel <- true
				}

				if len(coins) >= int(p.N-p.F) {
					coin, err := crypto.CombineSignatures(coins, ids)
					if err != nil {
						fmt.Println("error:", err)
					}
					l := utils.BytesToUint32(coin) % p.N // leader of round r
					thisRoundLeader <- l                 // for message handler
					leaderChannel <- l                   // for main process
					return
				}
			}
		}
	}()

	l := <-thisRoundLeader
	if p.PID == 1 {
		p.Log.Tracef("[p=%d,r=%d]: Leader is %d", p.PID, r, l)
	}

	// HaltMessage Handler
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Halt", IDr):
				payload := core.Decapsulation("Halt", m).(*protobuf.Halt)

				h := sha3.Sum512(payload.Value)
				var buf bytes.Buffer
				buf.Write([]byte("Echo"))
				buf.Write(IDrj[l])
				buf.WriteByte(2)
				buf.Write(h[:])
				sm := buf.Bytes()
				err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||l||2||h)
				if err == nil {
					haltChannel <- payload.Value
					if p.PID == 1 {
						p.Log.Tracef("[p=%d,r=%d]: Commit a value: receive a vaild Halt message", p.PID, r)
					}
					return
				}
			}
		}

	}()

	// TockCatBestMessage Handler
	go func() {
		BNr := [][]byte{}
		ids := []int64{}
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("TockCatBest", IDr):
				payload := core.Decapsulation("TockCatBest", m).(*protobuf.TockCatBest)
				if payload.Sigshare == nil {
					h := sha3.Sum512(payload.Value)
					var buf bytes.Buffer
					buf.Write([]byte("Echo"))
					buf.Write(IDrj[l])
					buf.WriteByte(1)
					buf.Write(h[:])
					sm := buf.Bytes()
					err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig1) //verify("Echo"||ID||r||l||h)
					if err != nil {
						p.Log.Warn("Invaild TockCatBest message: Sig1 is invalid")
						continue
					}

					h2 := sha3.Sum512(payload.Value)
					var buf2 bytes.Buffer
					buf2.Write([]byte("Echo"))
					buf2.Write(IDrj[l])
					buf2.WriteByte(2)
					buf2.Write(h2[:])
					sm2 := buf2.Bytes()
					err2 := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm2, payload.Sig2) //verify("Echo"||ID||r||2||h)
					if err2 != nil {
						p.Log.Warn("Invaild TockCatBest message: Sig2 is invalid")
						continue
					}
					Q2.Store(utils.BytesToUint32(payload.Leader), &protobuf.Finish{
						Value: payload.Value,
						Sig:   payload.Sig2,
					})

					voteFlagChannel <- 0
					voteYesChannel <- payload.Value
					voteYesChannel <- payload.Sig1
					return
				} else {
					var buf bytes.Buffer
					buf.Write([]byte("BAD LEADER"))
					buf.Write(IDr)
					sm := buf.Bytes()
					err := crypto.VerifyShare(p.Config.Keys.ThresholdPK, sm, payload.Sigshare)
					if err != nil {
						p.Log.Warn("Invaild TockCatBest message: Sigshare is invalid")
						continue
					}

					BNr = append(BNr, payload.Sigshare)
					ids = append(ids, int64(m.Sender)+1)
					if len(BNr) >= int(p.N-p.F) {
						bnSignature, err := crypto.CombineSignatures(BNr, ids)
						if err != nil {
							fmt.Println("error: ", err)
						}
						if p.PID == 1 {
							p.Log.Tracef("CombineSignatures for BAD LEADER")
						}

						voteFlagChannel <- 1
						voteNoChannel <- bnSignature
					}
				}
			}
		}
	}()
}
