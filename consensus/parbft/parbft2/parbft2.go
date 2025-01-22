package parbft2

import (
	"context"
	"time"

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
	var op1 *PathOutput
	op1 = nil
	finAgrTag := ""
	bd := false // 𝑏𝑑 indicates whether 𝑝𝑖 has broadcast data

	ctx, cancel := context.WithCancel(context.Background())
	optCtx, optCancel := context.WithCancel(ctx)   //son of ctx
	pessCtx, pessCancel := context.WithCancel(ctx) //son of ctx

	go OptPath2(optCtx, p, ID, value, leader, candidateChannel, commitChannel, op1)
	go func() {
		time.Sleep(time.Duration(p.TimeOut) * time.Millisecond) // wait until the timer of 5Δ expires
		bd = true
		if op1 != nil {
			if finAgrTag == "" {
				finAgrTag = op1.Tag
				go FinAgr(ctx, p, ID, op1, commitChannel)
				return
			}
		} else {
			go PessPath(pessCtx, p, ID, value, candidateChannel)

			for {
				select {
				case <-ctx.Done():
					return
				case candidate := <-candidateChannel:
					if finAgrTag == "" { // Make sure FinAgr is started only once
						finAgrTag = candidate.Tag
						if finAgrTag == "OPT" {
							pessCancel()
						} else {
							optCancel()
						}
						go FinAgr(ctx, p, ID, candidate, commitChannel)
						return
					}
				}
			}
		}
	}()

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
		case <-ctx.Done():
			return nil
		case res := <-commitChannel:
			if res.Tag == "OPT" && bd {
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
