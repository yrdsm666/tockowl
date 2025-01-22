package executor

import (
	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/types"
)

type Executor interface {
	CommitBlock(block types.Block, proof []byte, consensusID int64)
	VerifyTx(tx types.RawTransaction) bool
}

func BuildExecutor(cfg *config.ExecutorConfig, log *logrus.Entry) Executor {
	if cfg.Type == "local" {
		return NewLocalExecutor(cfg, log)
	} else if cfg.Type == "remote" {
		return NewRemoteExecutor(cfg, log)
	}
	log.WithField("type", cfg.Type).Warn("executor type error")
	return nil
}
