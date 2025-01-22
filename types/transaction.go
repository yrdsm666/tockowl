package types

import (
	"crypto/sha256"

	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/protobuf/proto"
)

type RawTransaction []byte

type TxHash [32]byte

type RawReceipt []byte

func (rtx RawTransaction) Hash() TxHash {
	return sha256.Sum256(rtx)
}

func (rtx RawTransaction) ToTx() (*pb.Transaction, error) {
	tx := new(pb.Transaction)
	err := proto.Unmarshal([]byte(rtx), tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func RawTxArrayToBytes(rtxs []RawTransaction) [][]byte {
	brtxs := make([][]byte, len(rtxs))
	for i, rtx := range rtxs {
		brtxs[i] = RawTransaction(rtx)
	}
	return brtxs
}

func RawTxArrayFromBytes(btxs [][]byte) []RawTransaction {
	txs := make([]RawTransaction, len(btxs))
	for i, btx := range btxs {
		txs[i] = RawTransaction(btx)
	}
	return txs
}

func BuildByteTx(t pb.TransactionType, payload []byte) []byte {
	tx := &pb.Transaction{
		Type:    t,
		Payload: payload,
	}
	brtx, err := proto.Marshal(tx)
	utils.PanicOnError(err)
	return brtx
}
