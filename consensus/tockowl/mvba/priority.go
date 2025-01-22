package mvba

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/yrdsm666/tockowl/crypto"

	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
)

// Calculate the priority of replica
func CalculateReplicaPriority(replicaID uint32, seed int) int {
	bytebuf := bytes.NewBuffer([]byte{})

	replicaIDBytes := utils.Uint32ToBytes(replicaID)
	seedBytes := utils.IntToBytes(seed)
	bytebuf.Write(replicaIDBytes)
	bytebuf.Write(seedBytes)

	priority := crypto.Hash(bytebuf.Bytes())

	return utils.BytesToInt(priority)
}

func MaxPriorityReplica(total uint32, seed int) uint32 {
	maxPriority := 0
	maxReplica := uint32(0)
	for i := uint32(0); i < total; i++ {
		priority := CalculateReplicaPriority(i, seed)
		if priority < 0 {
			fmt.Printf("Failed to calculate priority for replica %d, priority is %d.\n", i, priority)
		}

		if priority > maxPriority {
			maxPriority = priority
			maxReplica = i
		}
	}
	return maxReplica
}

func BestElement(valueSet *sync.Map, seed int, p *party.HonestParty, epoch uint32) (uint32, []byte, []byte) {
	var proposal []byte
	var sig []byte
	replicaID := uint32(0)
	maxPriority := 0

	// Find the leader
	if p.FastTrack {
		v, ok := valueSet.Load(p.GetLeaderByEpoch(epoch))
		if ok {
			lock, ok := (v).(*protobuf.Lock)
			if ok {
				proposal = lock.Value
				sig = lock.Sig
			} else {
				finish, ok := (v).(*protobuf.Finish)
				if ok {
					proposal = finish.Value
					sig = finish.Sig
				} else {
					fmt.Println("Failed to traverse sync.Map, value is invaild for leader.")
				}
			}

			if p.LightLoad {
				if len(proposal) > minProposalSize {
					return p.GetLeaderByEpoch(epoch), proposal, sig
				}
			}

		}
	}

	valueSet.Range(func(key, value interface{}) bool {
		k, ok := (key).(uint32)
		if !ok {
			fmt.Println("Failed to traverse sync.Map, key is invaild.")
			return false
		}
		priority := CalculateReplicaPriority(k, seed)
		if priority < 0 {
			fmt.Printf("Failed to calculate priority for replica %d, priority is %d.\n", k, priority)
		}
		var nowValue []byte
		var nowSig []byte

		if priority > maxPriority {

			lock, ok := (value).(*protobuf.Lock)
			if ok {
				nowValue = lock.Value
				nowSig = lock.Sig
			} else {
				finish, ok := (value).(*protobuf.Finish)
				if ok {
					nowValue = finish.Value
					nowSig = finish.Sig
				} else {
					fmt.Println("Failed to traverse sync.Map, value is invaild.")
				}
			}

			if p.LightLoad {
				// Confirm if the value is empty
				if len(nowValue) > minProposalSize {
					replicaID = k
					maxPriority = priority
					proposal = nowValue
					sig = nowSig
				}
				if len(nowValue) <= minProposalSize && maxPriority == 0 {
					replicaID = k
					maxPriority = 0
					proposal = nowValue
					sig = nowSig
				}
			} else {
				replicaID = k
				maxPriority = priority
				proposal = nowValue
				sig = nowSig
			}
		}
		return true
	})
	// fmt.Println("Max replicaID:", replicaID)
	return replicaID, proposal, sig
}
