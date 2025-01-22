package executor

import (
	"context"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/types"
	"github.com/yrdsm666/tockowl/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type LocalExecutor struct {
	counter int
	log     *logrus.Entry
}

func NewLocalExecutor(cfg *config.ExecutorConfig, uplog *logrus.Entry) *LocalExecutor {
	return &LocalExecutor{
		counter: 0,
		log:     uplog.WithField("app", "local executor"),
	}
}

func (e *LocalExecutor) CommitBlock(block types.Block, proof []byte, cid int64) {
	for _, rtx := range block.GetTxs() {
		tx, err := types.RawTransaction(rtx).ToTx()
		if tx.Type != pb.TransactionType_NORMAL {
			continue
		}
		utils.PanicOnError(err)
		split := strings.Split(string(tx.Payload), ",")
		arg1, _ := strconv.Atoi(split[0])
		arg2, _ := strconv.Atoi(split[1])
		e.counter++
		rawReceipt := []byte(strconv.Itoa(e.counter) + "--" + strconv.Itoa(arg1+arg2))
		msg := &pb.Msg{Payload: &pb.Msg_Reply{Reply: &pb.Reply{Tx: rtx, Receipt: rawReceipt}}}
		msgByte, err := proto.Marshal(msg)
		utils.PanicOnError(err)
		address := "localhost:9999"
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			e.log.Warn("connect to ", address, "failed")
			return
		}
		client := pb.NewP2PClient(conn)
		packet := &pb.Packet{
			Msg:         msgByte,
			ConsensusID: -1,
			Epoch:       -1,
			Type:        pb.PacketType_CLIENTPACKET,
		}

		if _, err = client.Send(context.Background(), packet); err != nil {
			e.log.Warn("reply to ", address, " failed")
		}
		conn.Close()
	}
}

func (e *LocalExecutor) VerifyTx(tx types.RawTransaction) bool {
	return true
}
