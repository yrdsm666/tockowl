package bkr

import (
	"bytes"
	"context"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/bkr/ba"
	"github.com/yrdsm666/tockowl/consensus/bkr/rbc"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"

	"github.com/yrdsm666/tockowl/consensus/pkg/utils"

	"github.com/shirou/gopsutil/mem"
)

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

func MainProcess(p *party.HonestParty, ID []byte, input []byte) [][]byte {
	// fmt.Printf("INPUT --- pid: %d, c: %s\n", p.PID, hex.EncodeToString(input)[:10])
	// to propose
	var buf bytes.Buffer
	buf.Write(ID)
	buf.Write(utils.Uint32ToBytes(p.PID))
	IDi := buf.Bytes()
	// Node initiates an RBC to broadcast its proposal
	go func() {
		if !p.IsByzantineNode(int64(p.PID)) {
			rbc.RBCSend(context.Background(), p, IDi, input)
		}
	}()

	// listen to RBC instances
	type RBCRes struct {
		ID     uint32
		Output []byte
	}
	var rbcChannelLock sync.Mutex
	rbcChannel := make(chan *RBCRes, p.N)
	rbcValues := make([][]byte, p.N)

	var baChannelLock sync.Mutex
	baChannel := make(chan *RBCRes, p.N)
	baValues := make([][]byte, p.N)

	hasVotedInBA := make([]bool, p.N)
	bkrValues := make([][]byte, p.N)

	var finishNodeLock sync.Mutex
	finishNode := uint32(0)      // finishNode records the number of nodes that have completed RBC and BA
	finishBAWithOne := uint32(0) // finishBAWithOne records the number of nodes that output 1 for the BA instance

	haltChannel := make(chan bool, p.N)

	ctx, cancel := context.WithCancel(context.Background())

	for i := uint32(0); i < p.N; i++ {
		var buf bytes.Buffer
		buf.Write(ID)
		buf.Write(utils.Uint32ToBytes(i))
		IDj := buf.Bytes()

		go func(j uint32) {
			// Receive RBC broadcast proposals from other nodes
			rbcOutput := rbc.RBCReceive(ctx, p, j, IDj)
			if rbcOutput == nil {
				return
			}

			if p.PID == 1 {
				p.Log.Tracef("[ pid = %d ]: Receive RBC from node %d", p.PID, j)
			}

			rbcRes := &RBCRes{
				ID:     j,
				Output: rbcOutput.Data,
			}
			rbcChannelLock.Lock()
			rbcChannel <- rbcRes
			rbcChannelLock.Unlock()

		}(i)
	}

	go func() {
		// Waiting for all RBCs to be output
		index := uint32(0)
		for {
			select {
			case <-ctx.Done():
				return
			case res := <-rbcChannel:
				rbcValues[res.ID] = res.Output

				// Update finishNode
				if len(baValues[res.ID]) != 0 {
					if baValues[res.ID][0] == 1 {
						bkrValues[res.ID] = res.Output
						finishNodeLock.Lock()
						finishNode++
						finishNodeLock.Unlock()
					}

				}
				// Confirm that the BA of all nodes has been output, and the RBC_j associated with BA_j of output 1 has also been output.
				// This condition can be triggered by either RBC output or BA output.
				if finishNode == p.N {
					haltChannel <- true
				}

				go func() {
					if !hasVotedInBA[res.ID] {
						// If input has not been provided to BA_j, input 1 to BA_j
						hasVotedInBA[res.ID] = true
						var buf bytes.Buffer
						buf.Write(ID)
						buf.Write(utils.Uint32ToBytes(res.ID))
						IDj := buf.Bytes()
						if p.PID == 1 {
							p.Log.Tracef("[ pid = %d ]: Init BA for node %d with 1", p.PID, res.ID)
						}
						var baOutput uint8
						if !p.IsByzantineNode(int64(p.PID)) {
							baOutput = ba.InitBA(p, IDj, []byte{1})
						} else {
							baOutput = ba.InitBA(p, IDj, []byte{0})
						}

						baRes := &RBCRes{
							ID:     res.ID,
							Output: []byte{baOutput},
						}

						baChannelLock.Lock()
						baChannel <- baRes
						baChannelLock.Unlock()
					}
				}()

			}
			index++
			if index >= p.N {
				break
			}
		}
	}()

	for i := uint32(0); i < p.N; i++ {
		res := <-baChannel

		baValues[res.ID] = res.Output
		if p.PID == 1 {
			p.Log.Tracef("[ pid = %d ]: BA for node %d output %d", p.PID, res.ID, res.Output[0])
		}

		// Update finishBAWithOne and finishNode
		finishNodeLock.Lock()
		if baValues[res.ID][0] == 1 {
			finishBAWithOne++
			// BA_j outputs 1, and needs to obtain the value of RBC_j
			if len(rbcValues[res.ID]) != 0 {
				// If RBC_j has been completed, the operations related to node j are considered to have been completed
				bkrValues[res.ID] = rbcValues[res.ID]
				finishNode++
			}
		} else {
			// BA_j outputs 0, no need to obtain the value of RBC_j
			finishNode++
		}
		// Confirm that the BA of all nodes has been output, and the RBC_j associated with BA_j of output 1 has also been output.
		// This condition can be triggered by either RBC output or BA output.
		if finishNode == p.N {
			haltChannel <- true
		}
		finishNodeLock.Unlock()

		// If enough BAs have output 1, input 0 to the BAs that have not yet input
		if finishBAWithOne == p.N-p.F {
			for j := uint32(0); j < p.N; j++ {
				go func(k uint32) {
					if !hasVotedInBA[k] {
						// If input has not been provided to BA_j, input 0 to BA_j
						hasVotedInBA[k] = true
						var buf bytes.Buffer
						buf.Write(ID)
						buf.Write(utils.Uint32ToBytes(k))
						IDk := buf.Bytes()
						if p.PID == 1 {
							p.Log.Tracef("[ pid = %d ]: Init BA for node %d with 0", p.PID, k)
						}

						baOutput := ba.InitBA(p, IDk, []byte{0})
						baRes := &RBCRes{
							ID:     k,
							Output: []byte{baOutput},
						}

						baChannelLock.Lock()
						baChannel <- baRes
						baChannelLock.Unlock()
					}
				}(j)
			}

		}
	}
	<-haltChannel
	cancel()
	return bkrValues
}
