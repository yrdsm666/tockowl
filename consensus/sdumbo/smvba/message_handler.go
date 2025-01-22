package smvba

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

func messageHandler(ctx context.Context, p *party.HonestParty, IDr []byte, IDrj [][]byte, Fr *sync.Map,
	doneFlagChannel chan bool,
	preVoteFlagChannel chan bool, preVoteYesChannel chan []byte, preVoteNoChannel chan []byte,
	voteFlagChannel chan byte, voteYesChannel chan []byte, voteNoChannel chan []byte, voteOtherChannel chan []byte,
	leaderChannel chan uint32, haltChannel chan []byte, r uint32) {

	//FinishMessage Handler
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
					Fr.Store(m.Sender, payload)
					FrLength++
					// if FrLength == int(2*p.F+1) {
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
					l := utils.BytesToUint32(coin) % p.N //leader of round r
					thisRoundLeader <- l                 //for message handler
					leaderChannel <- l                   //for main process
					return
				}
			}
		}
	}()

	l := <-thisRoundLeader

	//HaltMessage Handler
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
					return
				}
			}
		}

	}()

	//PreVoteMessage Handler
	go func() {
		PNr := [][]byte{}
		ids := []int64{}
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("PreVote", IDr):
				payload := core.Decapsulation("PreVote", m).(*protobuf.PreVote)
				if payload.Vote {
					h := sha3.Sum512(payload.Value)
					var buf bytes.Buffer
					buf.Write([]byte("Echo"))
					buf.Write(IDrj[l])
					buf.WriteByte(1)
					buf.Write(h[:])
					sm := buf.Bytes()
					err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||l||1||h)
					if err == nil {
						sm[len([]byte("Echo"))+len(IDrj[l])] = 2
						sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("Echo"||ID||r||l||2||h)
						preVoteFlagChannel <- true
						preVoteYesChannel <- payload.Value
						preVoteYesChannel <- payload.Sig
						preVoteYesChannel <- sigShare
					}
				} else {
					var buf bytes.Buffer
					buf.WriteByte(byte(0)) //false
					buf.Write(IDr)
					PNr = append(PNr, payload.Sig)
					ids = append(ids, int64(m.Sender)+1)
					if len(PNr) >= int(p.N-p.F) {
						noSignature, err := crypto.CombineSignatures(PNr, ids)
						if err != nil {
							fmt.Println("error: ", err)
							continue
						}
						var buf bytes.Buffer
						buf.Write([]byte("Unlock"))
						buf.Write(IDr)
						sm := buf.Bytes()
						sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("Unlock"||ID||r)
						preVoteFlagChannel <- false
						preVoteNoChannel <- noSignature
						preVoteNoChannel <- sigShare
					}
				}
			}
		}
	}()

	//VoteMessage Handler
	go func() {
		VYr := [][]byte{}
		IdV := []int64{}
		VNr := [][]byte{}
		IdN := []int64{}
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Vote", IDr):

				payload := core.Decapsulation("Vote", m).(*protobuf.Vote)
				if payload.Vote {
					h := sha3.Sum512(payload.Value)
					var buf bytes.Buffer
					buf.Write([]byte("Echo"))
					buf.Write(IDrj[l])
					buf.WriteByte(1)
					buf.Write(h[:])
					sm := buf.Bytes()
					err1 := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||l||1||h)
					sm[len([]byte("Echo"))+len(IDrj[l])] = 2
					if err1 == nil {
						VYr = append(VYr, payload.Sigshare)
						IdV = append(IdV, int64(m.Sender)+1)
						if len(VYr) >= int(p.N-p.F) {
							sig, err := crypto.CombineSignatures(VYr, IdV)
							if err != nil {
								fmt.Println("error: ", err)
							}
							voteFlagChannel <- 0
							voteYesChannel <- payload.Value
							voteYesChannel <- sig
						} else if len(VYr)+len(VNr) >= int(p.N-p.F) {
							voteFlagChannel <- 2
							voteOtherChannel <- payload.Value
							voteOtherChannel <- payload.Sig
						}
					}
				} else {
					var buf1 bytes.Buffer
					buf1.WriteByte(byte(0)) //false
					buf1.Write(IDr)
					sm1 := buf1.Bytes()
					err1 := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm1, payload.Sig) //verify(false||ID||r)

					var buf2 bytes.Buffer
					buf2.Reset()
					buf2.Write([]byte("Unlock"))
					buf2.Write(IDr)
					if err1 == nil {
						VNr = append(VNr, payload.Sigshare)
						IdN = append(IdN, int64(m.Sender)+1)
						if len(VNr) >= int(p.N-p.F) {
							sig, err := crypto.CombineSignatures(VNr, IdN)
							if err != nil {
								fmt.Println("error: ", err)
							}
							voteFlagChannel <- 1
							voteNoChannel <- sig
						}
					}
				}
			}
		}
	}()
}
