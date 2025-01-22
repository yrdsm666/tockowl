package tockcat

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

func NewTockCatConsensus(
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
		go tockcatTest(p)
	}

	return p
}

func tockcatTest(p *party.HonestParty) {
	epoch := p.Config.Epoch
	p.Metric.StartMetric()
	for k := uint32(0); k < uint32(epoch); k++ {

		ID := utils.Uint32ToBytes(k)
		// stx := fmt.Sprintf("%d ", p.ID)
		var tx []byte

		if !p.IsLightLoad(p.ID) {
			// tx = []byte(stx + "abcdefg")
			tx = utils.GetTxs(p.Config.BatchSize, p.Config.PayloadSize)
		} else {
			// tx = []byte(stx)
			tx = []byte{}
		}

		startTime := time.Now()

		ctx, cancel := context.WithCancel(context.Background())
		res := MainProcess(ctx, p, ID, tx, nil, Q)
		cancel()

		elapsedTime := time.Since(startTime)
		p.Metric.LatencyMeasurement(elapsedTime)
		p.Metric.ThroughputMeasurement(res)
		p.Metric.AddTotalEpoch()

		fmt.Printf("[TOCKCAT MVBA] Epoch %d: GOOD WORK for replica %d: %s\n", k, p.ID, hex.EncodeToString(res)[:10])
	}
}

func Q(p *party.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
	return nil
}
