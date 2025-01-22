package acs

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"log"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/tockowl/mvba"
	"github.com/yrdsm666/tockowl/consensus/tockowl/pb"
	"github.com/yrdsm666/tockowl/crypto"

	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/reedsolomon"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/consensus/pkg/vectorcommitment"

	"github.com/shirou/gopsutil/mem"
	"github.com/vivint/infectious"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

func MainProcess(p *party.HonestParty, epoch []byte, input []byte) [][]byte {
	// Store proposals
	var pStore = store{
		b:     [][]bool{},
		data:  [][][]byte{},
		mutex: new(sync.Mutex)}
	pStore.grow(p.N)

	// Propose
	var buf bytes.Buffer
	buf.Write(epoch)
	buf.Write(utils.Uint32ToBytes(p.PID))
	IDi := buf.Bytes()
	// Run this party's PB instance
	go func() {
		h, sig, ok := pb.Sender(context.Background(), p, IDi, input, nil)
		if ok {
			// Broadcast lock message
			lockMessage := core.Encapsulation("BLock", IDi, p.PID, &protobuf.BLock{
				Hash: h,
				Sig:  sig,
			})
			p.Log.Trace("Broadcast BLock message")
			p.Broadcast(lockMessage)
		}
	}()

	// Listen to PB instances
	var mutex sync.Mutex
	lockPIDChannel := make(chan uint32, p.N)
	lockHashChannel := make(chan []byte, p.N)
	lockSigChannel := make(chan []byte, p.N)
	for i := uint32(0); i < p.N; i++ {
		var buf bytes.Buffer
		buf.Write(epoch)
		buf.Write(utils.Uint32ToBytes(i))
		IDj := buf.Bytes()

		// Handle lock
		go func(j uint32) {
			// Receive proposal
			proposal, _, _ := pb.Receiver(context.Background(), p, j, IDj, nil, nil, nil)
			pStore.store(1, j, proposal)
			m := <-p.GetMessage("BLock", IDj)
			payload := core.Decapsulation("BLock", m).(*protobuf.BLock)
			var buf bytes.Buffer
			buf.Write([]byte("Echo"))
			buf.Write(IDj)
			buf.Write(payload.Hash)
			sm := buf.Bytes()
			err := crypto.VerifySignature(p.Config.Keys.ThresholdPK, sm, payload.Sig) //verify("Echo"||e||j||h)
			h := sha3.Sum512(proposal)
			if err == nil && bytes.Equal(payload.Hash, h[:]) {
				mutex.Lock() //important
				lockPIDChannel <- j
				lockHashChannel <- payload.Hash
				lockSigChannel <- payload.Sig
				mutex.Unlock()
				// p.Log.Trace("Receive BLock message")
			}

		}(i)
	}

	// Wait to invoke MVBA
	pids := []uint32{}
	hashes := [][]byte{}
	sigs := [][]byte{}
	for i := uint32(0); i < p.N-p.F; i++ {
		pids = append(pids, <-lockPIDChannel)
		hashes = append(hashes, <-lockHashChannel)
		sigs = append(sigs, <-lockSigChannel)
	}
	value, err1 := proto.Marshal(&protobuf.BLockSetValue{
		Pid:  pids,
		Hash: hashes,
	})
	validation, err2 := proto.Marshal(&protobuf.BLockSetValidation{
		Sig: sigs,
	})
	if err1 != nil || err2 != nil {
		log.Fatalln(err1, err2)
	}

	// Wait for MVBA's output
	p.Log.Trace("Start mvba")
	ctx, cancel := context.WithCancel(context.Background())
	resultValue := mvba.MainProcess(ctx, p, epoch, value, validation, Q)
	cancel()
	p.Log.Trace("The result of mvba: ", hex.EncodeToString(resultValue))
	var L protobuf.BLockSetValue //L={(j,h)}
	proto.Unmarshal(resultValue, &L)

	go helper(p, epoch, &pStore)

	// Call help
	S := []uint32{}
	for i := uint32(0); i < p.N-p.F; i++ {
		ok := pStore.isStored(1, L.Pid[i])
		if !ok {
			S = append(S, L.Pid[i])
		}
	}
	callMessage := core.Encapsulation("Call", epoch, p.PID, &protobuf.Call{
		Indices: S,
	})

	for i := uint32(0); i < p.N; i++ {
		if i == p.PID {
			continue
		}
		p.Unicast(p.GetNetworkInfo()[int64(i)], callMessage)
	}

	// Recovery
	shares := make([][]infectious.Share, p.N)
	haltChannel := make(chan bool, 1)
	coder := reedsolomon.NewRScoder(int(p.F+1), int(p.N))
	doneNum := 0
	if len(S) > 0 {
	doneFlag:
		for {
			select {
			case <-haltChannel:
				break doneFlag

			case m := <-p.GetMessage("Help", epoch):
				payload := core.Decapsulation("Help", m).(*protobuf.Help)

				for i, index := range payload.Indices {
					shardAndProof := payload.ShardAndProof[i]
					if vectorcommitment.VerifyMerkleTreeProof(shardAndProof.Vc, shardAndProof.Proof1, shardAndProof.Proof2, shardAndProof.Shard) {
						shares[index] = append(shares[index], infectious.Share{
							Data:   shardAndProof.Shard,
							Number: int(m.Sender),
						})
						if len(shares[index]) == int(p.F+1) {
							value, err := coder.Decode(shares[index]) //decode
							if err != nil {
								panic(err)
							}
							pStore.store(1, index, value)
							doneNum++
							if doneNum == len(S) {
								haltChannel <- true
								break
							}
						}
					}
				}
			}
		}
	}

	output := [][]byte{}
	for _, index := range L.Pid {
		v, _ := pStore.load(1, index)
		output = append(output, v)
	}
	return output
}

func helper(p *party.HonestParty, epoch []byte, pStore *store) {
	// Listen to others' call
	go func() {
		coder := reedsolomon.NewRScoder(int(p.F+1), int(p.N))
		for {
			m := <-p.GetMessage("Call", epoch)
			payload := (core.Decapsulation("Call", m)).(*protobuf.Call)

			shardIndices := []uint32{}
			shards := [][]byte{}
			vcs := [][]byte{}
			proof1s := [][][]byte{}
			proof2s := [][]int64{}
			for _, index := range payload.Indices {
				element, ok := pStore.load(1, index)
				if ok {
					//shardIndices
					shardIndices = append(shardIndices, index)

					//shards
					temp := coder.Encode(element)
					fragments := make([][]byte, p.N)
					for i := uint32(0); i < p.N; i++ {
						fragments[i] = temp[i].Data
					}
					shards = append(shards, fragments[p.PID])

					//vcs
					vCommiter, _ := vectorcommitment.NewMerkleTree(fragments)
					vc := vCommiter.GetMerkleTreeRoot()
					vcs = append(vcs, vc)

					//proof1s, proof2s
					proof1, proof2 := vCommiter.GetMerkleTreeProof(int(p.PID))
					proof1s = append(proof1s, proof1)
					proof2s = append(proof2s, proof2)

				}
			}

			shardAndProof := []*protobuf.Help_ShardAndProof{}
			for i := range shardIndices {
				shardAndProof = append(shardAndProof, &protobuf.Help_ShardAndProof{
					Vc:     vcs[i],
					Shard:  shards[i],
					Proof1: proof1s[i],
					Proof2: proof2s[i],
				})
			}

			if len(shardIndices) > 0 {
				helpMessage := core.Encapsulation("Help", epoch, p.PID, &protobuf.Help{
					Indices:       shardIndices,
					ShardAndProof: shardAndProof,
				})
				p.Unicast(p.GetNetworkInfo()[int64(m.Sender)], helpMessage)
			}

		}
	}()
}

func Q(p *party.HonestParty, epoch []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
	var L protobuf.BLockSetValue //L={(j,h)}
	proto.Unmarshal(value, &L)

	var S protobuf.BLockSetValidation
	proto.Unmarshal(validation, &S)

	if len(L.Hash) != int(p.N-p.F) || len(L.Pid) != int(p.N-p.F) || len(S.Sig) != int(p.N-p.F) {
		return errors.New("Q check failed")
	}

	for i := uint32(0); i < p.N-p.F; i++ {
		h, ok1 := hashVerifyMap.Load(L.Pid[i])
		s, ok2 := sigVerifyMap.Load(L.Pid[i])
		if ok1 && ok2 {
			if bytes.Equal(L.Hash[i], h.([]byte)) && bytes.Equal(S.Sig[i], s.([]byte)) {
				continue
			} else {
				return nil
			}
		}
		var buf bytes.Buffer
		buf.Write([]byte("Echo"))
		buf.Write(epoch[:4])
		buf.Write(utils.Uint32ToBytes(L.Pid[i]))
		buf.Write(L.Hash[i])
		sm := buf.Bytes()
		err := crypto.VerifyShare(p.Config.Keys.ThresholdPK, sm, S.Sig[i]) //verify("Echo"||e||j||h)
		if err != nil {
			return err
		}
		hashVerifyMap.Store(L.Pid[i], L.Hash[i])
		sigVerifyMap.Store(L.Pid[i], S.Sig[i])
	}

	return nil
}
