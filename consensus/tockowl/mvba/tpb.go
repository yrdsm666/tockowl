package mvba

import (
	"bytes"
	"context"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/crypto"
	"golang.org/x/crypto/sha3"
)

// strong provable broadcast
func tpbSender(ctx context.Context, p *party.HonestParty, ID []byte, value []byte, validation []byte, parentQc1 *protobuf.Lock) ([]byte, []byte, bool) {
	var buf1, buf2, buf3 bytes.Buffer
	buf1.Write(ID)
	buf1.WriteByte(1)
	buf2.Write(ID)
	buf2.WriteByte(2)
	buf3.Write(ID)
	buf3.WriteByte(3)
	ID1 := buf1.Bytes()
	ID2 := buf2.Bytes()
	ID3 := buf3.Bytes()

	// safeProposal check is only performed on the first vote
	_, sig1, ok1 := Sender(ctx, p, ID1, value, validation, parentQc1)
	if ok1 {
		// Byzantine replicas only perform the first-phase of broadcast
		if p.IsByzantineNode(int64(p.PID)) {
			return value, nil, false
		}

		_, sig2, ok2 := Sender(ctx, p, ID2, value, sig1, nil)
		if ok2 {
			_, sig3, ok3 := Sender(ctx, p, ID3, value, sig2, nil)
			if ok3 {
				return value, sig3, true //FINISH
			}
		}
	}
	return nil, nil, false

}

func tpbReceiver(ctx context.Context, p *party.HonestParty, sender uint32, ID []byte,
	V *sync.Map, Q1 *sync.Map, Q2 *sync.Map,
	validator1 func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error,
	safeProposal func(uint32, uint32, int) bool, parentQc2Owner uint32, preSeed int,
	hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) bool {

	var buf1, buf2, buf3 bytes.Buffer
	buf1.Write(ID)
	buf1.WriteByte(1)
	buf2.Write(ID)
	buf2.WriteByte(2)
	buf3.Write(ID)
	buf3.WriteByte(3)
	ID1 := buf1.Bytes()
	ID2 := buf2.Bytes()
	ID3 := buf3.Bytes()

	// safeProposal check is only performed on the first vote
	value0, validation, ok0 := Receiver(ctx, p, sender, ID1, validator1, safeProposal, parentQc2Owner, preSeed, hashVerifyMap, sigVerifyMap)
	if !ok0 {
		return false
	}
	V.Store(sender, &protobuf.Lock{
		Value: value0,
		Sig:   validation,
	})

	value1, sig1, ok1 := Receiver(ctx, p, sender, ID2, validator2, nil, 0, 0, hashVerifyMap, sigVerifyMap)
	if !ok1 {
		return false
	}
	Q1.Store(sender, &protobuf.Lock{
		Value: value1,
		Sig:   sig1,
	})

	value2, sig2, ok2 := Receiver(ctx, p, sender, ID3, validator3, nil, 0, 0, hashVerifyMap, sigVerifyMap)
	if !ok2 {
		return false
	}
	Q2.Store(sender, &protobuf.Lock{
		Value: value2,
		Sig:   sig2,
	})

	return true //LOCK
}

func validator2(p *party.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap, sigVerifyMap *sync.Map) error {
	h := sha3.Sum512(value)
	var buf bytes.Buffer
	buf.Write([]byte("Echo"))
	buf.Write(ID[:len(ID)-1])
	buf.WriteByte(1)
	buf.Write(h[:])
	sm := buf.Bytes()
	// err := bls.Verify(pairing.NewSuiteBn256(), p.SigPK.Commit(), sm, validation)
	err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, validation)
	return err
}

func validator3(p *party.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap, sigVerifyMap *sync.Map) error {
	h := sha3.Sum512(value)
	var buf bytes.Buffer
	buf.Write([]byte("Echo"))
	buf.Write(ID[:len(ID)-1])
	buf.WriteByte(2)
	buf.Write(h[:])
	sm := buf.Bytes()
	// err := bls.Verify(pairing.NewSuiteBn256(), p.SigPK.Commit(), sm, validation)
	err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, validation)
	return err
}
