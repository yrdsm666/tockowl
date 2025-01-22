package ba

import (
	"context"
	"encoding/hex"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
)

type BA struct {
	ID               []byte
	p                *party.HonestParty
	nowRound         uint32
	BARoundinstances map[uint32]*BARoundinstance
	roundLock        sync.Mutex
	waitNewRound     *sync.Cond

	zeroFinish   uint32
	oneFinish    uint32
	hasVotedStop bool // Whether STOP has been sent
	resChannel   chan uint8
}

// ID = epoch + owner
func InitBA(p *party.HonestParty, ID []byte, input []byte) uint8 {
	// fmt.Printf("INPUT --- pid: %d, c: %d\n", p.PID, input)
	ctx, cancel := context.WithCancel(context.Background())
	ba := &BA{
		ID:               ID,
		p:                p,
		nowRound:         0,
		BARoundinstances: make(map[uint32]*BARoundinstance),
		resChannel:       make(chan uint8),
	}
	ba.waitNewRound = sync.NewCond(&ba.roundLock)

	// STOP message hander
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("STOP", ID):
				payload := core.Decapsulation("STOP", m).(*protobuf.STOP)
				switch payload.Content[0] {
				case 0:
					ba.zeroFinish++
				case 1:
					ba.oneFinish++
				}
				// Stop voting
				if !ba.hasVotedStop && ba.zeroFinish > p.F {
					ba.sendSTOP(0)
				}
				if !ba.hasVotedStop && ba.oneFinish > p.F {
					ba.sendSTOP(1)
				}

				if ba.zeroFinish >= p.N-p.F {
					// End
					ba.p.Log.Tracef("[r=%d, pid=%d] RES: %d ", ba.nowRound, ba.p.PID, 0)
					ba.resChannel <- 0
					return
				}
				if ba.oneFinish >= p.N-p.F {
					// End
					ba.p.Log.Tracef("[r=%d, pid=%d] RES: %d ", ba.nowRound, ba.p.PID, 1)
					ba.resChannel <- 1
					return
				}
			}
		}
	}(ctx)

	// new round for BA
	ba.newBARoundinstance(input)

	// message hander
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("BVAL", ID):
				payload := core.Decapsulation("BVAL", m).(*protobuf.BVAL)
				round := payload.Round
				if round <= 0 {
					continue
				}
				go func() {
					// get current round
					ba.roundLock.Lock()
					bari := ba.expectNewRound(round)
					ba.roundLock.Unlock()
					bari.lock.Lock()
					defer bari.lock.Unlock()

					bari.handleBVALMessage(payload)
					if bari.p.PID == 1 {
						bari.p.Log.Tracef("[r=%d, pid=%d] Receive BVAL(%d) from %d", bari.round, bari.p.PID, payload.Content[0], m.Sender)
					}

				}()

			case m := <-p.GetMessage("AUX", ID):
				payload := core.Decapsulation("AUX", m).(*protobuf.AUX)
				round := payload.Round
				if round <= 0 {
					continue
				}
				go func() {
					ba.roundLock.Lock()
					bari := ba.expectNewRound(round)
					ba.roundLock.Unlock()
					bari.lock.Lock()
					defer bari.lock.Unlock()

					bari.handleAUXMessage(payload)
					if bari.p.PID == 1 {
						bari.p.Log.Tracef("[r=%d, pid=%d] Receive AUX ", bari.round, bari.p.PID)
					}
				}()

			case m := <-p.GetMessage("COIN", ID):
				payload := core.Decapsulation("COIN", m).(*protobuf.COIN)
				round := payload.Round
				if round <= 0 {
					continue
				}
				go func() {
					ba.roundLock.Lock()
					bari := ba.expectNewRound(round)
					ba.roundLock.Unlock()
					if bari.p.PID == 1 {
						bari.p.Log.Tracef("[r=%d, pid=%d] Receive COIN", bari.round, bari.p.PID)
					}

					bari.lock.Lock()
					defer bari.lock.Unlock()

					bari.coinSigs = append(bari.coinSigs, payload.Content)
					bari.coinMsgSenders = append(bari.coinMsgSenders, int64(m.Sender)+1)
					bari.isReadyToEnterNewRound()
				}()

			}
		}
	}(ctx)

	// Waiting for the result
	result := <-ba.resChannel
	cancel()

	return result
}

func (ba *BA) newBARoundinstance(value []byte) {
	// var buf bytes.Buffer
	// // buf.Write(epoch)
	// buf.Write(utils.Uint32ToBytes(round))
	// IDr := buf.Bytes()

	ba.nowRound++
	if ba.p.PID == 1 {
		ba.p.Log.Tracef("====== round %d, value %d ======", ba.nowRound, value)
	}
	if ba.nowRound >= 20 {
		ba.p.Log.Warn(" BAD WORK IN BA ")
		return
	}

	bari := &BARoundinstance{
		ID:             ba.ID,
		p:              ba.p,
		round:          ba.nowRound,
		coinSigs:       make([][]byte, 0),
		coinMsgSenders: make([]int64, 0),
		ba:             ba,
	}
	ba.roundLock.Lock()
	ba.BARoundinstances[ba.nowRound] = bari
	ba.roundLock.Unlock()
	ba.waitNewRound.Broadcast()

	BVALMessage := core.Encapsulation("BVAL", ba.ID, ba.p.PID, &protobuf.BVAL{
		// Owner:   owner,
		Round:   bari.round,
		Content: value,
	})
	err := ba.p.Broadcast(BVALMessage)
	if err != nil {
		ba.p.Log.Warnf("[party %v] multicast BVALMessage error: %v, BA instance ID: %s", ba.p.PID, err, hex.EncodeToString(ba.ID))
	}
	ba.p.Log.Tracef("[r=%d, pid=%d] Broadcast BVAL ", bari.round, ba.p.PID)
	switch value[0] {
	case 0:
		bari.hasVotedZero = true
	case 1:
		bari.hasVotedOne = true
	default:
		bari.p.Log.Warn("valueBytes: ", value, ", valueBytes[0]: ", value[0])
	}
}

func (ba *BA) sendSTOP(value uint8) {
	if !ba.hasVotedStop {
		ba.p.Log.Trace("send STOP in round: ", ba.nowRound)
		err := ba.p.Broadcast(core.Encapsulation("STOP", ba.ID, ba.p.PID, &protobuf.STOP{
			// Owner:   owner,
			Content: []byte{value},
		}))
		if err != nil {
			ba.p.Log.Warnf("[party %v] multicast STOPMessage error: %v, BA instance ID: %s", ba.p.PID, err, hex.EncodeToString(ba.ID))
		}
		ba.hasVotedStop = true
	}
}

func (ba *BA) expectNewRound(round uint32) *BARoundinstance {
	for {
		bari, ok := ba.BARoundinstances[round]
		if !ok {
			ba.waitNewRound.Wait()
		} else {
			return bari
		}
	}
}
