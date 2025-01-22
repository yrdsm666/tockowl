package aab

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/tockowl/acs"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"

	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/crypto"

	"github.com/shirou/gopsutil/mem"
)

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

func TockowlAcsTest(p *party.HonestParty, outputChannel chan []byte) {
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

		output := []byte{}
		for _, v := range pResult {
			p.Metric.ThroughputMeasurement(v)
			output = append(output, v...)
		}

		fmt.Printf("[TOCKOWL ACS] Epoch %d: GOOD WORK for replica %d: %s\n", e, p.ID, hex.EncodeToString(output)[:10])
	}
}
