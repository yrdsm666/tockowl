package fin

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yrdsm666/tockowl/config"
	"github.com/yrdsm666/tockowl/consensus/fin/acs"
	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"
	"github.com/yrdsm666/tockowl/executor"
	"github.com/yrdsm666/tockowl/p2p"
)

func NewFinConsensus(
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
		go finTest(p, nil)
	}
	return p
}

func finTest(p *party.HonestParty, outputChannel chan []byte) {
	epoch := p.Config.Epoch
	p.Metric.StartMetric()
	for e := uint32(0); e < uint32(epoch); e++ {

		tx := utils.GetTxs(p.Config.BatchSize, p.Config.PayloadSize)

		startTime := time.Now()

		c, err := crypto.ThresholdEncrypt(p.Config.Keys.EncPK, tx)
		if err != nil {
			log.Fatalln(err)
		}

		// Invoke ACS
		ID := utils.Uint32ToBytes(e)
		cResult := acs.MainProcess(p, ID, c)
		for i, c := range cResult {
			if len(c) == 0 {
				continue
			}
			//fmt.Printf("FIN --- pid: %d, id: %d, c: %s\n", p.PID, i, hex.EncodeToString(c)[:10])
			decShare, err := crypto.ThresholdDecShare(p.Config.Keys.EncSK, c)
			if err != nil {
				log.Fatalln(err)
			}

			decMessage := core.Encapsulation("Dec", ID, p.PID, &protobuf.Dec{
				Id:       uint32(i),
				DecShare: decShare,
			})
			p.Broadcast(decMessage)
		}

		pResult := make([][]byte, len(cResult))
		decShares := make([][][]byte, len(cResult))

		doneNum := 0
		for {
			m := <-p.GetMessage("Dec", ID)
			payload := core.Decapsulation("Dec", m).(*protobuf.Dec)
			if pResult[payload.Id] != nil {
				continue
			}
			decShares[payload.Id] = append(decShares[payload.Id], payload.DecShare)
			if len(decShares[payload.Id]) == int(p.F+1) {
				pt, err := crypto.ThresholdDecrypt(p.Config.Keys.EncVK, cResult[payload.Id], decShares[payload.Id], int(p.F+1), int(p.N))
				if err != nil {
					log.Fatalln(err)
					break
				} else {
					pResult[payload.Id] = pt
					doneNum++
				}
			}

			if doneNum == len(cResult) {
				break
			}
		}

		elapsedTime := time.Since(startTime)
		p.Metric.LatencyMeasurement(elapsedTime)
		p.Metric.AddTotalEpoch()

		// output := []byte{}
		logRes := make([]string, len(pResult))
		for i := 0; i < len(logRes); i++ {
			if len(pResult[i]) == 0 {
				logRes[i] = ""
				continue
			}
			p.Metric.ThroughputMeasurement(pResult[i])
			logRes[i] = hex.EncodeToString(pResult[i])[:10]
		}

		fmt.Printf("[FIN ACS] Epoch %d: GOOD WORK for replica %d: %s\n", e, p.ID, logRes)
		// outputChannel <- output
	}
}
