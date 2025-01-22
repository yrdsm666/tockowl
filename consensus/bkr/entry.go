package bkr

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/bkr/ba"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"

	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

func NewBAConsensus(
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
		go baTest(p)
	}

	return p
}

func NewBKRConsensus(
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
		go bkrTest(p)
	}

	return p
}

func baTest(p *party.HonestParty) {
	epoch := p.Config.Epoch
	// epoch := 1
	p.Metric.StartMetric()

	for k := uint32(0); k < uint32(epoch); k++ {
		ID := utils.Uint32ToBytes(k)
		seed := rand.NewSource(time.Now().UnixNano())
		r := rand.New(seed)
		input := uint8(r.Intn(2))

		// input := []byte{num}

		startTime := time.Now()
		res := ba.InitBA(p, ID, []byte{input})

		p.Log.Infof("================ BA: Epoch = %d , Input = %d , Output = %d ================\n", k, input, res)
		if p.PID == 1 {
			fmt.Println("Number of coroutines:", runtime.NumGoroutine())
		}

		elapsedTime := time.Since(startTime)
		p.Metric.LatencyMeasurement(elapsedTime)
		// p.Metric.ThroughputMeasurement([]byte{res})
		p.Metric.AddTotalEpoch()

		// fmt.Printf("Epoch %d: GOOD WORK for replica %d: %s\n", k, p.ID, string(res))
	}
}

func bkrTest(p *party.HonestParty) {
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

		res := MainProcess(p, ID, tx)

		logRes := make([]string, len(res))
		for i := uint32(0); i < p.N; i++ {
			if len(res[i]) == 0 {
				logRes[i] = ""
				continue
			}
			p.Metric.ThroughputMeasurement(res[i])
			logRes[i] = hex.EncodeToString(res[i])[:10]
			// fmt.Printf("BKR output --- pid: %d, id: %d, c: %s\n", p.PID, i, hex.EncodeToString(res[i])[:10])
		}
		fmt.Printf("BKR output in round %d: --- pid: %d, c: %s\n", k, p.PID, logRes)

		elapsedTime := time.Since(startTime)
		p.Metric.LatencyMeasurement(elapsedTime)
		p.Metric.AddTotalEpoch()

		// fmt.Printf("Epoch %d: GOOD WORK for replica %d: %s\n", k, p.ID, string(res))
	}
}
