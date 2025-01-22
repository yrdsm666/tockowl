package parbft2

import (
	"bytes"
	"context"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"
	"golang.org/x/crypto/sha3"
)

// MainProcess is the main process of smvba instances
func OptPath2(ctx context.Context, p *party.HonestParty, ID []byte, input []byte, leader uint32,
	candidateChannel chan *PathOutput, commitChannel chan *PathOutput, optCandidateIsOk *PathOutput) {

	hashVerifyMap := sync.Map{}
	sigVerifyMap := sync.Map{}

	//Initialize SPB instances
	var buf bytes.Buffer
	buf.Write(ID)
	buf.Write(utils.Uint32ToBytes(leader))
	IDl := buf.Bytes()

	var validator func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error
	validator = nil // TODO
	go func() {
		value, sig, ok := spbReceiver(ctx, p, leader, IDl, validator, &hashVerifyMap, &sigVerifyMap)
		if ok { //save Lock
			optCandidate := &PathOutput{
				Tag:   "OPT",
				Value: value,
				Sig:   sig,
			}
			candidateChannel <- optCandidate
			optCandidateIsOk = optCandidate
		}
	}()

	go func() {
		if p.PID == leader {
			value, sig, ok := spbSender(ctx, p, IDl, input, nil)
			if ok {
				finishMessage := core.Encapsulation("Finish", ID, p.PID, &protobuf.Finish{
					Value: value,
					Sig:   sig,
				})
				p.Broadcast(finishMessage)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-p.GetMessage("Finish", ID):
			payload := core.Decapsulation("Finish", m).(*protobuf.Finish)
			h := sha3.Sum512(payload.Value)
			var buf bytes.Buffer
			buf.Write([]byte("Echo"))
			buf.Write(IDl)
			buf.WriteByte(2)
			buf.Write(h[:])
			sm := buf.Bytes()
			// err := bls.Verify(pairing.NewSuiteBn256(), p.SigPK.Commit(), sm, payload.Sig) //verify("Echo"||ID||r||j||2||h)
			err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||ID||r||j||2||h)
			if err == nil {
				optCommit := &PathOutput{
					Tag:   "OPT",
					Value: payload.Value,
					Sig:   payload.Sig,
				}
				if p.PID == 1 {
					p.Log.Tracef("[p=%d]: Commit a value in OptPath1", p.PID)
				}
				commitChannel <- optCommit
			}
		}
	}
}
