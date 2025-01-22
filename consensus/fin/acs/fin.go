package acs

import (
	"bytes"
	"context"
	"encoding/hex"
	"log"
	"sync"

	// "github.com/yrdsm666/tockowl/consensus/bkr/rbc"
	"github.com/yrdsm666/tockowl/consensus/bkr/rbc"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/tockowl/mvba"

	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"

	"github.com/shirou/gopsutil/mem"
	"google.golang.org/protobuf/proto"
)

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

// return format: [sig0, sig1, null, sig3]
func MainProcess(p *party.HonestParty, epoch []byte, input []byte) [][]byte {
	// fmt.Printf("INPUT --- pid: %d, c: %s\n", p.PID, hex.EncodeToString(input))
	// Propose
	var buf bytes.Buffer
	buf.Write(epoch)
	buf.Write(utils.Uint32ToBytes(p.PID))
	IDi := buf.Bytes()
	// Run this party's RBC instance
	go func() {
		rbc.RBCSend(context.Background(), p, IDi, input)
	}()

	// Listen to RBC instances
	var mutex sync.Mutex
	type RBCRes struct {
		ID     uint32
		Output []byte
	}
	valueChannel := make(chan *RBCRes, p.N)

	for i := uint32(0); i < p.N; i++ {
		var buf bytes.Buffer
		buf.Write(epoch)
		buf.Write(utils.Uint32ToBytes(i))
		IDj := buf.Bytes()

		// Handle lock
		go func(j uint32) {
			// Receive proposal
			output := rbc.RBCReceive(context.Background(), p, j, IDj).Data
			res := &RBCRes{
				ID:     j,
				Output: output,
			}
			mutex.Lock()
			valueChannel <- res
			mutex.Unlock()
		}(i)
	}

	// Wait to invoke MVBA
	values := make([][]byte, p.N)
	for i := uint32(0); i < p.N-p.F; i++ {
		res := <-valueChannel
		values[res.ID] = res.Output
	}
	value, err1 := proto.Marshal(&protobuf.BLockSetValidation{
		Sig: values,
	})

	if err1 != nil {
		log.Fatalln(err1)
	}

	// Wait for MVBA's output
	p.Log.Tracef("FIN: MVBA input value: %s\n", hex.EncodeToString(value)[:10])
	ctx, cancel := context.WithCancel(context.Background())
	resultValue := mvba.MainProcess(ctx, p, epoch, value, nil, Q)
	cancel()
	p.Log.Trace("FIN: The result of mvba: ", hex.EncodeToString(resultValue)[:10])
	var L protobuf.BLockSetValidation
	proto.Unmarshal(resultValue, &L)

	output := [][]byte{}
	for _, v := range L.Sig {
		if v != nil && len(hex.EncodeToString(v)) > 0 {
			output = append(output, v)
		}
	}

	return output
}

func Q(p *party.HonestParty, epoch []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
	return nil
}
