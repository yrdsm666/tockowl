package parbft2

import (
	"bytes"
	"context"
	"log"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/bkr/ba"
	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/crypto"
	"golang.org/x/crypto/sha3"
)

func FinAgr(ctx context.Context, p *party.HonestParty, ID []byte, input *PathOutput, commitChannel chan *PathOutput) {
	vals := make([]*PathOutput, 2)

	h := sha3.Sum512(input.Value)
	var buf bytes.Buffer
	buf.Write([]byte("FinAgr"))
	buf.Write([]byte(input.Tag))
	buf.Write(h[:])
	sm := buf.Bytes()
	sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("FinAgr"||Tag||h)

	prepMessage := core.Encapsulation("PREP", ID, p.PID, &protobuf.PREP{
		Tag:      []byte(input.Tag),
		Data:     input.Value,
		Sig:      input.Sig,
		Sigshare: sigShare,
	})
	p.Broadcast(prepMessage)

	sigs := [][]byte{}
	ids := []int64{}
	optPrepNum := 0                   // The number of PERP messages with tag OPT
	pessPrepNum := 0                  // The number of PERP messages with tag PESS
	optPrepRecord := &protobuf.PREP{} // Record the content of a PERP message with tag OPT
	var baLock sync.Mutex
	baInput := uint8(0)
	baOutput := uint8(0)
	baFinish := false

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("PREP", ID):
				payload := core.Decapsulation("PREP", m).(*protobuf.PREP)
				if bytes.Equal(payload.Tag, []byte("OPT")) {
					err := crypto.VerifyShare(p.Config.Keys.VerifyKeys[int(m.Sender)], sm, payload.Sigshare) //verify("Echo"||e||j||h)
					if err != nil {
						p.Log.Fatalln(err)
						continue
					}
					sigs = append(sigs, payload.Sigshare)
					ids = append(ids, int64(m.Sender)+1)
					optPrepNum++
					if optPrepNum == 1 {
						// Record the first PERP message with tag OPT
						optPrepRecord = payload
					}
				} else {
					pessPrepNum++
				}

				if optPrepNum+pessPrepNum >= int(p.N-p.F) {
					if optPrepNum >= int(p.N-p.F) {
						// All PERP messages have OPT tag
						signature, err := crypto.CombineSignatures(sigs, ids)
						if err != nil {
							log.Fatalln(err)
						}

						optCommit := &PathOutput{
							Tag:   "OPT",
							Value: payload.Data,
							Sig:   signature,
						}

						if p.PID == 1 {
							p.Log.Tracef("[p=%d]: Commit a value in FinAgr: n-f PREP messages with tag OPT", p.PID)
						}
						commitChannel <- optCommit
						return
					} else if optPrepNum > 0 {
						// At least one PERP message has the tag PESS
						faMessage := core.Encapsulation("FA", ID, p.PID, &protobuf.FA{
							Tag:  []byte("OPT"),
							Data: optPrepRecord.Data,
							Sig:  optPrepRecord.Sig,
						})
						vals[0] = &PathOutput{
							Value: optPrepRecord.Data,
							Sig:   optPrepRecord.Sig,
						}
						p.Broadcast(faMessage)
						baInput = 0
					} else {
						// All PERP messages have PESS tag
						faMessage := core.Encapsulation("FA", ID, p.PID, &protobuf.FA{
							Tag:  []byte("PESS"),
							Data: payload.Data,
							Sig:  payload.Sig,
						})
						vals[1] = &PathOutput{
							Value: payload.Data,
							Sig:   payload.Sig,
						}
						p.Broadcast(faMessage)
						baInput = 1
					}

					// Waiting for the output of ABA
					baRes := ba.InitBA(p, ID, []byte{baInput})

					baLock.Lock()
					baOutput = baRes
					baFinish = true

					// If ABA outputs baOutput and vals[baOutput] is not empty, the node can commit
					if vals[baOutput] != nil {
						if p.PID == 1 {
							p.Log.Tracef("[p=%d]: Commit a value in FinAgr: BA output a value", p.PID)
						}
						commitChannel <- vals[baOutput]
						baLock.Unlock()
						return
					}

					baLock.Unlock()
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("FA", ID):
				payload := core.Decapsulation("FA", m).(*protobuf.FA)
				if bytes.Equal(payload.Tag, []byte("OPT")) {
					if vals[0] == nil {
						vals[0] = &PathOutput{
							Value: payload.Data,
							Sig:   payload.Sig,
						}
					}
				} else {
					if vals[1] == nil {
						vals[1] = &PathOutput{
							Value: payload.Data,
							Sig:   payload.Sig,
						}
					}
				}
				baLock.Lock()
				// If ABA outputs baOutput and vals[baOutput] is not empty, the node can commit
				if baFinish {
					if vals[baOutput] != nil {
						if p.PID == 1 {
							p.Log.Tracef("[p=%d]: Commit a value in FinAgr: receive a FA message", p.PID)
						}
						commitChannel <- vals[baOutput]
						baLock.Unlock()
						return
					}
				}
				baLock.Unlock()
			}
		}
	}()
}
