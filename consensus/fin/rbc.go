package fin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/klauspost/reedsolomon"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"google.golang.org/protobuf/proto"
)

// received message shard (RS code) in RBC
type mReceived struct {
	j  int
	mj []byte
}

type hm struct {
	h string
	m string
}

// RBC propose
func RBCSend(ctx context.Context, p *party.HonestParty, ID []byte, input []byte) {
	M := &protobuf.Message{Type: "RBCData", Id: ID, Sender: p.PID, Data: input}
	//encapsulate
	data, _ := proto.Marshal(M)
	err := p.Broadcast(&protobuf.Message{Type: "RBCPropose", Id: ID, Sender: p.PID, Data: data})
	if err != nil {
		p.Log.Tracef("[party %v] multicast RBCPropose error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
	}
	p.Log.Tracef("[RBC][Party %v] broadcast RBCPropose message, instance ID: %s\n", p.PID, hex.EncodeToString(ID))
}

func RBCReceive(ctx context.Context, p *party.HonestParty, sender uint32, ID []byte) *protobuf.Message {
	var M1 = make([][]byte, p.N) //  M' in the RBC paper (line 9, Algo 4). Must assign length (or copy will fail)
	var mLen int                 // the length of M_i
	//here we ignore the validator function P(.)

	//map (hash,M) to counter. Key value doesn't support []byte, so we transform it to string type.
	var EchoMessageMap = make(map[string]map[string]int)

	//T_h in line 16, Algo 4, RBC paper. T maps the hash to []m_received = {(index,mj), ...}
	var T = make(map[string][]mReceived)
	var MaxReadyNumber = int(0)
	var ReadyNumberChannel = make(chan int, p.N)
	var MaxReadyHash []byte

	var isReadySent = false
	var mutex sync.Mutex // isReadySent will be written by two goroutines. (line 11 and 13, Algo 4)
	var mutexEchoMap sync.Mutex
	var mutexReadyMap sync.Mutex

	var readyChan = make(chan hm, p.N)
	var RSDecStartChan = make(chan bool, 1)
	var RSDecStart = false

	var hasReceivedReady = make(map[uint32]bool) //avoid replay attack
	for i := uint32(0); i < p.N; i++ {
		hasReceivedReady[i] = false
	}

	// handle RBCPropose message
	go func() {
		m := <-p.GetMessage("RBCPropose", ID)
		p.Log.Tracef("[party %v] receive RBCPropose message from [party %v], RBC instance ID: %s", p.PID, m.Sender, hex.EncodeToString(ID))
		M := m.Data
		mLen = len(M) // the length of M is used to remove the padding zeros after RS decoding
		hLocal := sha256.New()
		hLocal.Write(M)

		//encode
		RSEncoder, _ := reedsolomon.New(int(p.N-(p.F+1)), int(p.F+1))
		shards, _ := RSEncoder.Split(M)
		RSEncoder.Encode(shards)
		//copy, avoid "M1" becoming nullPointer when "shards" is released at the end of this goroutine
		copy(M1, shards)
		for j := uint32(0); j < p.N; j++ {
			//encapsulate, append the length of M at the end of hash
			EchoData, _ := proto.Marshal(&protobuf.RBCEcho{Hash: append(hLocal.Sum(nil), utils.IntToBytes(mLen)...), M: shards[j]})
			p.Log.Tracef("[party %v] sends RBCEcho to [party %v], address : %s, RBC instance ID: %s", p.PID, j, p.GetNetworkInfo()[int64(j)], hex.EncodeToString(ID))
			if j == p.PID {
				p.Unicast(p.GetSelfInfo().Address, &protobuf.Message{Type: "RBCEcho", Sender: p.PID, Id: ID, Data: EchoData})
			} else {
				p.Unicast(p.GetNetworkInfo()[int64(j)], &protobuf.Message{Type: "RBCEcho", Sender: p.PID, Id: ID, Data: EchoData})
			}

			/*
				// This block is used for RBC_test when a party does not receive 2t+1 ECHO messages
				// In this case, the algorithm sends RBCReady message through line 15 upon receiving
				// t+1 RBCReady messages and t+1 matching RBCEcho messages
				if p.PID == 0 || p.PID == 1 {
					if index != uint32(2) {
						p.Send(&protobuf.Message{Type: "RBCEcho", Sender: p.PID, Id: ID, Data: EchoData}, index)
					}
				} else {
					p.Send(&protobuf.Message{Type: "RBCEcho", Sender: p.PID, Id: ID, Data: EchoData}, index)
				}
			*/
		}

	}()

	//handle Echo Message, line 11-12, Algo 4 in RBC paper
	go func(ctx context.Context) {
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("RBCEcho", ID):
				p.Log.Tracef("[party %v] receive RBCEcho message from [party %v], RBC instance ID: %s", p.PID, m.Sender, hex.EncodeToString(ID))
				var payloadMessage protobuf.RBCEcho
				proto.Unmarshal(m.Data, &payloadMessage)
				hash := string(payloadMessage.Hash)
				mi := string(payloadMessage.M)

				mutexEchoMap.Lock()
				_, ok1 := EchoMessageMap[hash]
				if ok1 {
					//ok1 denotes that the map of hash exists
					counter, ok2 := EchoMessageMap[hash][mi]
					if ok2 {
						//ok2 denotes that the map of (hash,M) exists, then increase the counter
						EchoMessageMap[hash][mi] = counter + 1
					} else {
						//else establish the map of (hash,M) and set it as 1
						EchoMessageMap[hash][mi] = 1
					}
				} else {
					//else establish the map of (hash,M) and set it as 1
					EchoMessageMap[hash] = make(map[string]int)
					EchoMessageMap[hash][mi] = 1
				}

				//send RBCReady, upon receiving n-t=2f+1 matching RBCEcho messages and not having sent RBCReady (line 11-12, Algo 4)
				mutex.Lock()
				if !isReadySent && uint32(EchoMessageMap[hash][mi]) >= p.N-p.F {
					isReadySent = true
					readyData, _ := proto.Marshal(&protobuf.RBCReady{Hash: []byte(hash), M: []byte(mi)})
					err := p.Broadcast(&protobuf.Message{Type: "RBCReady", Sender: p.PID, Id: ID, Data: readyData})
					if err != nil {
						p.Log.Tracef("[party %v] multicast RBCReady message error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
					}
					p.Log.Tracef("[party %v] multicast RBCReady in the 2t+1 Echo case, RBC instance ID: %s", p.PID, hex.EncodeToString(ID))
				}

				if uint32(EchoMessageMap[hash][mi]) >= p.F+1 {
					//ready to send RBCReady when receive t+1 RBCReady messages from the others
					p.Log.Tracef("[party %v] has received t+1 RBCEcho messages, send to readyChan, RBC instance ID:%s", p.PID, hex.EncodeToString(ID))
					readyChan <- hm{hash, mi}
				}
				mutex.Unlock() //TODO: move this unlock into if statement
				mutexEchoMap.Unlock()
				//finish this goroutine when RBCReady is sent
				if isReadySent {
					break loop
				}
			}
		}
	}(ctx)

	//handle RBCReady messages
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-p.GetMessage("RBCReady", ID):
				if hasReceivedReady[m.Sender] {
					continue
				} else {
					hasReceivedReady[m.Sender] = true
				}
				p.Log.Tracef("[party %v] receive RBCReady message from [party %v], RBC instance ID: %s", p.PID, m.Sender, hex.EncodeToString(ID))
				var payloadMessage protobuf.RBCReady
				proto.Unmarshal(m.Data, &payloadMessage)
				hash := payloadMessage.Hash
				hashString := string(hash)
				mJ := payloadMessage.M
				j := m.Sender //sender's pid, i.e., the index of the sender

				mutexReadyMap.Lock()
				//avoid the replay attack

				//
				_, ok := T[hashString]
				if ok {
					T[hashString] = append(T[hashString], mReceived{int(j), mJ})
				} else {
					T[hashString] = make([]mReceived, 0)
					T[hashString] = append(T[hashString], mReceived{int(j), mJ})
				}

				if len(T[hashString]) > MaxReadyNumber {
					MaxReadyNumber = len(T[hashString])
					MaxReadyHash = hash
					ReadyNumberChannel <- MaxReadyNumber
				}

				//send RBCReady messages, line 13-15, Algo 4 in RBC paper
				mutex.Lock()
				if !isReadySent && uint32(len(T[hashString])) >= p.F+1 {
					mutex.Unlock()
					p.Log.Tracef("[party %v] receive t+1 RBCReady message, RBC instance ID: %s\n", p.PID, hex.EncodeToString(ID))
					//wait for t+1 matching RBCEcho messages
					var res hm
					for {
						select {
						case <-ctx.Done():
							return
						case res = <-readyChan:
							p.Log.Tracef("[party %v] waiting for t+1 matching RBCEcho messages, RBC instance ID: %s", p.PID, hex.EncodeToString(ID))
							mutexEchoMap.Lock()
							if uint32(EchoMessageMap[res.h][res.m]) >= p.F+1 {
								mutexEchoMap.Unlock()
								p.Log.Tracef("[party %v] finish waiting for t+1 matching RBCEcho messages, RBC instance ID: %s", p.PID, hex.EncodeToString(ID))
								mutex.Lock()
								isReadySent = true
								mutex.Unlock()
								readyData, _ := proto.Marshal(&protobuf.RBCReady{Hash: hash, M: []byte(res.m)})
								err := p.Broadcast(&protobuf.Message{Type: "RBCReady", Sender: p.PID, Id: ID, Data: readyData})
								if err != nil {
									p.Log.Tracef("[party %v] multicast RBCReady error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
								}
								p.Log.Tracef("[party %v] multicast RBCReady in the t+1 Ready case, RBC instance ID: %s\n", p.PID, hex.EncodeToString(ID))
								break
							}
							mutexEchoMap.Unlock()
						}
					}
				} else {
					mutex.Unlock()
				}

				// upon receiving |T| >= 2t+1 = n-t
				if !RSDecStart && uint32(len(T[string(hash)])) >= p.N-p.F {
					RSDecStart = true
					RSDecStartChan <- true
				}
				mutexReadyMap.Unlock()

				//FIXME: kill this for loop when the RBC message is reconstructed
				// isRec := <-RecDone
				// if isRec {
				// 	break
				// }
			}
		}
	}(ctx)

	// wait for at least 2t+1 = n-t RS shards in T_h, i.e., T[string(hash)]
	select {
	case <-ctx.Done():
		return nil
	case <-RSDecStartChan:
		p.Log.Tracef("[party %v] start reconstruction, RBC instance ID: %s\n", p.PID, hex.EncodeToString(ID))
		for r := uint32(0); r <= p.F; r++ {
		loop:
			for {
				select {
				case <-ctx.Done():
					break loop
				case mrn := <-ReadyNumberChannel:
					if uint32(mrn) >= p.N-p.F+r {
						break loop
					}
				}
			}
			// for {
			// 	if uint32(MaxReadyNumber) >= p.N-p.F+r {
			// 		break
			// 	}
			// }

			var mReceivedTemp = make([]mReceived, p.N-p.F+r)
			mutexReadyMap.Lock()
			copy(mReceivedTemp, T[string(MaxReadyHash)])
			mutexReadyMap.Unlock()

			var M = make([][]byte, p.N)
			for i := uint32(0); i < p.N-p.F+r; i++ {
				M[mReceivedTemp[i].j] = mReceivedTemp[i].mj
			}
			RSEncoder, _ := reedsolomon.New(int(p.N-(p.F+1)), int(p.F+1))
			ok, _ := RSEncoder.Verify(M)
			if !ok {
				RSEncoder.Reconstruct(M)
			}

			//parse M and remove the padding zeros
			var mReconstructed = make([]byte, 0)
			for i := uint32(0); i < p.N-(p.F+1); i++ {
				mReconstructed = append(mReconstructed, M[i]...)
			}

			//the last several bytes in MaxReadyHash are the lenth of M' (see line 9, Algo 4 in RBC paper)
			mLenNew := utils.BytesToInt(MaxReadyHash[256/8:])
			//the first 256/8 bytes in MaxReadyHash are the hash value
			MaxReadyHash = MaxReadyHash[:256/8]
			mReconstructed = mReconstructed[0:mLenNew]

			hNew := sha256.New()
			hNew.Write(mReconstructed)
			if bytes.Equal(hNew.Sum(nil), MaxReadyHash) {
				var replyMessage protobuf.Message
				proto.Unmarshal(mReconstructed, &replyMessage)
				p.Log.Tracef("[party %v] have reconstructed message, RBC instance ID: %s\n", p.PID, hex.EncodeToString(ID))
				return &replyMessage
			}
		}
	}
	p.Log.Tracef("ERROR: [party %v] cannot reconstruct the message, RBC instance ID: %s", p.PID, hex.EncodeToString(ID))
	return nil
}
