package parbft1

import (
	"context"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"

	// "github.com/yrdsm666/tockowl/consensus/tockowl/mvba"
	"github.com/yrdsm666/tockowl/consensus/sdumbo/smvba"
)

type PathOutput struct {
	Tag   string
	Value []byte
	Sig   []byte
}

// MainProcess is the main process of parbft instances
func MainProcess(p *party.HonestParty, ID []byte, leader uint32, value []byte, validation []byte) []byte {
	candidateChannel := make(chan *PathOutput)
	commitChannel := make(chan *PathOutput)
	finAgrTag := ""

	ctx, cancel := context.WithCancel(context.Background())
	optCtx, optCancel := context.WithCancel(ctx)   //son of ctx
	pessCtx, pessCancel := context.WithCancel(ctx) //son of ctx

	go OptPath1(optCtx, p, ID, value, leader, candidateChannel, commitChannel)
	go PessPath(pessCtx, p, ID, value, candidateChannel)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("Halt", ID):
				payload := core.Decapsulation("Halt", m).(*protobuf.Halt)
				// TODO: Verify the validity of the halt message
				haltCommit := &PathOutput{
					Tag:   "OPT",
					Value: payload.Value,
					Sig:   payload.Sig,
				}
				if p.PID == 1 {
					p.Log.Tracef("[p=%d]: Commit a value in MainProcess: receive a HALT message", p.PID)
				}
				commitChannel <- haltCommit
			}
		}
	}()

	for {
		select {
		case candidate := <-candidateChannel:
			if finAgrTag == "" { // Make sure FinAgr is started only once
				finAgrTag = candidate.Tag
				if finAgrTag == "OPT" {
					pessCancel()
				} else {
					optCancel()
				}
				go FinAgr(ctx, p, ID, candidate, commitChannel)
			}
		case res := <-commitChannel:
			if res.Tag == "OPT" {
				haltMessage := core.Encapsulation("Halt", ID, p.PID, &protobuf.Halt{
					Value: res.Value,
					Sig:   res.Sig,
				})
				p.Broadcast(haltMessage)
			}

			cancel()
			return res.Value
		}
	}
}

func PessPath(ctx context.Context, p *party.HonestParty, epoch []byte, input []byte, candidateChannel chan *PathOutput) {
	// TockCat
	// pessOutput := mvba.MainProcess(ctx, p, epoch, input, nil, nil)

	// sMVBA
	pessOutput := smvba.MainProcess(ctx, p, epoch, input, nil, nil)
	result := &PathOutput{
		Tag:   "PESS",
		Value: pessOutput,
	}

	candidateChannel <- result
}
