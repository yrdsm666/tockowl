package parbft

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/parbft/parbft1"
	"github.com/yrdsm666/tockowl/consensus/parbft/parbft2"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

func NewParbftConsensus(
	id int64,
	cid int64,
	cfg *config.ConsensusConfig,
	exec executor.Executor,
	p2pAdaptor p2p.P2PAdaptor,
	log *logrus.Entry,
	partition string,
) *party.HonestParty {
	p := party.InitHonestParty(id, cid, cfg, exec, p2pAdaptor, log, partition)

	time.Sleep(500 * time.Millisecond)
	if !p.IsCrashNode(p.ID) {
		// Node is not a crashed node
		go ParbftTest(p)
	}

	return p
}

func ParbftTest(p *party.HonestParty) {
	epoch := p.Config.Epoch
	p.Metric.StartMetric()
	for e := uint32(0); e < uint32(epoch); e++ {
		ID := utils.Uint32ToBytes(e)
		leader := e % p.N
		tx := utils.GetTxs(p.Config.BatchSize, p.Config.PayloadSize)

		startTime := time.Now()

		var res []byte
		if p.Config.Parbft.Type == "parbft1" {
			res = parbft1.MainProcess(p, ID, uint32(leader), tx, nil)
		} else if p.Config.Parbft.Type == "parbft2" {
			res = parbft2.MainProcess(p, ID, uint32(leader), tx, nil)
		} else {
			p.Log.Warnf("Parbft consensus type not supported: %s", p.Config.Parbft.Type)
			return
		}

		elapsedTime := time.Since(startTime)
		p.Metric.LatencyMeasurement(elapsedTime)
		p.Metric.AddTotalEpoch()
		p.Metric.ThroughputMeasurement(res)

		// p.Log.Infof("================ BA: Epoch = %d , Output = %s ================\n", e, hex.EncodeToString(res)[:10])
		fmt.Printf("[PARBFT %s] Epoch %d: GOOD WORK for replica %d: %s\n", p.Config.Parbft.Type, e, p.ID, hex.EncodeToString(res)[:10])
	}
}
