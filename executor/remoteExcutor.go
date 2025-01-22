package executor

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/pb"
	"github.com/yrdsm666/tockowl/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RemoteExecutor struct {
	log     *logrus.Entry
	address string
	client  pb.ExecutorClient
}

func NewRemoteExecutor(cfg *config.ExecutorConfig, uplog *logrus.Entry) *RemoteExecutor {
	log := uplog.WithField("app", "local executor")
	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Warn("connect to ", cfg.Address, "failed")
	}
	client := pb.NewExecutorClient(conn)
	return &RemoteExecutor{
		log:     log,
		address: cfg.Address,
		client:  client,
	}
}

func (e *RemoteExecutor) CommitBlock(block types.Block, proof []byte, cid int64) {
	eb := &pb.ExecBlock{
		Txs: block.GetTxs(),
	}
	if _, err := e.client.CommitBlock(context.Background(), eb); err != nil {
		e.log.WithError(err).Warn("commit block failed")
	}
}

func (e *RemoteExecutor) VerifyTx(rtx types.RawTransaction) bool {
	tx, err := rtx.ToTx()
	if err != nil {
		e.log.WithError(err).Warn("convert to tx failed")
		return false
	}
	if result, err := e.client.VerifyTx(context.Background(), tx); err != nil {
		e.log.WithError(err).Warn("convert to tx failed")
		//TODO: change to false when deploy
		return true
	} else {
		return result.GetSuccess()
	}
}
