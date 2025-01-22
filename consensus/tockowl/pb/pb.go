package pb //provable broadcast

import (
	"bytes"
	"context"
	"log"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/crypto"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/protobuf/proto"

	"golang.org/x/crypto/sha3"
)

// Sender is run by the sender of a instance of provable broadcast
func Sender(ctx context.Context, p *party.HonestParty, ID []byte, value []byte, validation []byte) ([]byte, []byte, bool) {
	valueMessage := core.Encapsulation("Value", ID, p.PID, &protobuf.Value{
		Value:      value,
		Validation: validation,
	})

	p.Broadcast(valueMessage)

	sigs := [][]byte{}
	ids := []int64{}
	h := sha3.Sum512(value)
	var buf bytes.Buffer
	buf.Write([]byte("Echo"))
	buf.Write(ID)
	buf.Write(h[:])
	sm := buf.Bytes()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, false
		case m := <-p.GetMessage("Echo", ID):
			payload := core.Decapsulation("Echo", m).(*protobuf.Echo)
			err := crypto.VerifyShare(p.Config.Keys.VerifyKeys[int(m.Sender)], sm, payload.Sigshare) //verify("Echo"||e||j||h)
			if err != nil {
				log.Fatalln(err)
				continue
			}
			sigs = append(sigs, payload.Sigshare)
			ids = append(ids, int64(m.Sender)+1)

			// if len(sigs) > int(2*p.F) {
			if len(sigs) >= int(p.N-p.F) {
				// signature, err := tbls.Recover(pairing.NewSuiteBn256(), p.SigPK, sm, sigs, int(2*p.F+1), int(p.N))
				signature, err := crypto.CombineSignatures(sigs, ids)
				if err != nil {
					log.Fatalln(err)
				}

				return h[:], signature, true
			}
		}
	}

}

// Receiver is run by the receiver of a instance of provable broadcast
func Receiver(ctx context.Context, p *party.HonestParty, sender uint32, ID []byte, validator func(*party.HonestParty, []byte, []byte, []byte, *sync.Map, *sync.Map) error, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) ([]byte, []byte, bool) {
	select {
	case <-ctx.Done():
		return nil, nil, false
	case m := <-p.GetMessage("Value", ID):
		payload := (core.Decapsulation("Value", m)).(*protobuf.Value)
		if validator != nil {
			err := validator(p, ID, payload.Value, payload.Validation, hashVerifyMap, sigVerifyMap)
			if err != nil {
				log.Fatalln(err, "PB validator for", m.Sender)
				return nil, nil, false //sender is dishonest
			}
		}
		h := sha3.Sum512(payload.Value)
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write(ID)
		buf.Write(h[:])
		sm := buf.Bytes()
		sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("Echo"||ID||h)

		echoMessage := core.Encapsulation("Echo", ID, p.PID, &protobuf.Echo{
			Sigshare: sigShare,
		})
		if sender == p.PID {
			msgByte, err := proto.Marshal(echoMessage)
			utils.PanicOnError(err)
			p.MsgByteEntrance <- msgByte
		} else {
			p.Unicast(p.GetNetworkInfo()[int64(sender)], echoMessage)
		}

		return payload.Value, payload.Validation, true
	}
}
