package dumbo_ng

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

func NewDumboNgConsensus(
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
	// if !p.IsCrashNode(p.ID) {
	// Node is not a crashed node
	go dumboNgTest(p)
	// }

	return p
}

func dumboNgTest(p *party.HonestParty) {
	epoch := p.Config.Epoch

	ctx, cancel := context.WithCancel(context.Background())
	outputChannel := make(chan [][]byte, 1024)

	p.Metric.StartMetric()

	go MainProgress(ctx, p, outputChannel)

	// resultLen := 0
	for e := 0; e < epoch; e++ {

		values := <-outputChannel

		for i := 0; i < len(values); i++ {
			txWithTime := DecodeBytesToTxWithTime(values[i])
			elapsedTime := time.Since(txWithTime.Time)
			// if p.PID == 1 && i == 0 {
			// 	fmt.Println("time: ", elapsedTime, ", startTime: ", txWithTime.Time, ", endTime: ", time.Now())
			// }
			p.Metric.LatencyMeasurement(elapsedTime)
			p.Metric.ThroughputMeasurement(txWithTime.Txs)
		}
		p.Metric.AddTotalEpoch()

		fmt.Printf("[DUMBO-NG] Epoch %d: GOOD WORK for replica %d\n", e, p.ID)
	}

	cancel()

}
