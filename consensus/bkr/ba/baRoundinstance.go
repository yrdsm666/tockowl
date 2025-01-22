package ba

import (
	"bytes"
	"encoding/hex"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"
)

type BARoundinstance struct {
	ID    []byte
	p     *party.HonestParty
	round uint32
	ba    *BA

	hasVotedZero   bool // Whether 0 has been cast
	hasVotedOne    bool // Whether 1 has been cast
	hasSentAux     bool
	hasSentCoin    bool
	zeroEndorsed   bool
	oneEndorsed    bool
	lock           sync.Mutex
	isEnd          bool
	binVals        uint8
	numBvalZero    uint32 // The number of nodes that voted 0
	numBvalOne     uint32 // The number of nodes that voted 1
	numAuxZero     uint32
	numAuxOne      uint32
	coinSigs       [][]byte
	coinMsgSenders []int64
}

func (bari *BARoundinstance) handleBVALMessage(msg *protobuf.BVAL) {
	switch msg.Content[0] {
	case 0:
		bari.numBvalZero++
	case 1:
		bari.numBvalOne++
	default:
		bari.p.Log.Warn("Invaild BVAL messgae: ", msg.Content)
	}

	// Second BVAL voting
	if !bari.hasVotedZero && bari.numBvalZero > bari.p.F {
		err := bari.p.Broadcast(core.Encapsulation("BVAL", bari.ID, bari.p.PID, &protobuf.BVAL{
			// Owner:   owner,
			Round:   bari.round,
			Content: []byte{0},
		}))
		if err != nil {
			bari.p.Log.Warnf("[party %v] multicast BVALMessage error: %v, BA instance ID: %s", bari.p.PID, err, hex.EncodeToString(bari.ID))
		}
		bari.hasVotedZero = true
	}
	if !bari.hasVotedOne && bari.numBvalOne > bari.p.F {
		err := bari.p.Broadcast(core.Encapsulation("BVAL", bari.ID, bari.p.PID, &protobuf.BVAL{
			// Owner:   owner,
			Round:   bari.round,
			Content: []byte{1},
		}))
		if err != nil {
			bari.p.Log.Warnf("[party %v] multicast BVALMessage error: %v, BA instance ID: %s", bari.p.PID, err, hex.EncodeToString(bari.ID))
		}
		bari.hasVotedOne = true
	}

	// AUX step
	if !bari.zeroEndorsed && bari.numBvalZero >= bari.p.N-bari.p.F {
		// AUX vote(0)
		bari.zeroEndorsed = true
		if !bari.hasSentAux {
			bari.hasSentAux = true
			err := bari.p.Broadcast(core.Encapsulation("AUX", bari.ID, bari.p.PID, &protobuf.AUX{
				// Owner:   owner,
				Round:   bari.round,
				Content: []byte{0},
			}))
			if err != nil {
				bari.p.Log.Warnf("[party %v] multicast AUXMessage error: %v, BA instance ID: %s", bari.p.PID, err, hex.EncodeToString(bari.ID))
			}
			if bari.p.PID == 1 {
				bari.p.Log.Tracef("[r=%d, pid=%d] Broadcast AUX ", bari.round, bari.p.PID)
			}
		}
		bari.isReadyToSendCoin()
		bari.isReadyToEnterNewRound()
	}

	if !bari.oneEndorsed && bari.numBvalOne >= bari.p.N-bari.p.F {
		// AUX vote(1)
		bari.oneEndorsed = true
		if !bari.hasSentAux {
			bari.hasSentAux = true
			err := bari.p.Broadcast(core.Encapsulation("AUX", bari.ID, bari.p.PID, &protobuf.AUX{
				// Owner:   owner,
				Round:   bari.round,
				Content: []byte{1},
			}))
			if err != nil {
				bari.p.Log.Warnf("[party %v] multicast AUXMessage error: %v, BA instance ID: %s", bari.p.PID, err, hex.EncodeToString(bari.ID))
			}
			if bari.p.PID == 1 {
				bari.p.Log.Tracef("[r=%d, pid=%d] Broadcast AUX ", bari.round, bari.p.PID)
			}
		}
		bari.isReadyToSendCoin()
		bari.isReadyToEnterNewRound()
	}
}

func (bari *BARoundinstance) handleAUXMessage(msg *protobuf.AUX) {
	switch msg.Content[0] {
	case 0:
		bari.numAuxZero++
	case 1:
		bari.numAuxOne++
	default:
		bari.p.Log.Warn("Invaild AUX messgae")
	}

	bari.isReadyToSendCoin()
	bari.isReadyToEnterNewRound()
}

// must be executed within ba.lock
func (bari *BARoundinstance) isReadyToSendCoin() {
	if !bari.hasSentCoin {
		if bari.oneEndorsed && bari.numAuxOne >= bari.p.N-bari.p.F {
			if !bari.isEnd {
				bari.binVals = 1
			}
		} else if bari.zeroEndorsed && bari.numAuxZero >= bari.p.N-bari.p.F {
			if !bari.isEnd {
				bari.binVals = 0
			}
		} else if !bari.isEnd && bari.oneEndorsed && bari.zeroEndorsed &&
			bari.numAuxOne+bari.numAuxZero >= bari.p.N-bari.p.F {
			if !bari.isEnd {
				bari.binVals = 2
			}
		} else {
			return
		}
		bari.hasSentCoin = true
		sigShare := crypto.ThresholdSign(bari.p.Config.Keys.SecretKey, bari.getCoinInfo())
		coinMsg := &protobuf.COIN{
			// Owner:   owner,
			Round:   bari.round,
			Content: sigShare,
		}
		msg := core.Encapsulation("COIN", bari.ID, bari.p.PID, coinMsg)
		err := bari.p.Broadcast(msg)
		if err != nil {
			bari.p.Log.Warnf("[party %v] multicast COINMessage error: %v, BA instance ID: %s", bari.p.PID, err, hex.EncodeToString(bari.ID))
		}
		if bari.p.PID == 1 {
			bari.p.Log.Tracef("[r=%d, pid=%d] Broadcast COIN %d and sig %s", bari.round, bari.p.PID, coinMsg.Round, hex.EncodeToString(sigShare)[:10])
		}
	}
}

// must be executed within ba.lock
// return true if the instance is decided or finished at the first time
func (bari *BARoundinstance) isReadyToEnterNewRound() {
	if bari.hasSentCoin && !bari.isEnd &&
		len(bari.coinSigs) >= int(bari.p.N-bari.p.F) &&
		bari.numAuxZero+bari.numAuxOne >= bari.p.N-bari.p.F &&
		((bari.oneEndorsed && bari.numAuxOne >= bari.p.N-bari.p.F) ||
			(bari.zeroEndorsed && bari.numAuxZero >= bari.p.N-bari.p.F) ||
			(bari.oneEndorsed && bari.zeroEndorsed)) {
		sig, err := crypto.CombineSignatures(bari.coinSigs, bari.coinMsgSenders)
		if err != nil {
			bari.p.Log.Fatalln(err)
		}

		coin := sig[0] % 2
		bari.isEnd = true

		var nextValue byte
		if coin == bari.binVals {
			nextValue = bari.binVals
			bari.ba.sendSTOP(nextValue)
		} else if bari.binVals != 2 { // nextVote should insist the single value
			nextValue = bari.binVals
		} else {
			nextValue = coin
		}
		if bari.p.PID == 1 {
			bari.p.Log.Tracef("[r=%d, pid=%d] COIN: %d, bari.binVals: %d, nextValue: %d", bari.round, bari.p.PID, coin, bari.binVals, nextValue)
		}

		bari.ba.newBARoundinstance([]byte{nextValue})
	}
}

func (bari *BARoundinstance) getCoinInfo() []byte {
	var buf bytes.Buffer
	buf.Write([]byte("BA_COIN"))
	buf.Write(bari.ID)
	buf.Write(utils.Uint32ToBytes(bari.round))
	sm := buf.Bytes()

	return sm
}
