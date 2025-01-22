package rbc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"google.golang.org/protobuf/proto"
)

// RBC propose
func RBCSend(ctx context.Context, p *party.HonestParty, ID []byte, input []byte) {
	M := &protobuf.Message{Type: "RBCData", Id: ID, Sender: p.PID, Data: input}
	data, _ := proto.Marshal(M)
	err := p.Broadcast(&protobuf.Message{Type: "RBCPropose", Id: ID, Sender: p.PID, Data: data})
	if err != nil {
		p.Log.Tracef("[party %v] multicast RBCPropose error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
	}
	p.Log.Tracef("[RBC][Party %v] broadcast RBCPropose message, instance ID: %s\n", p.PID, hex.EncodeToString(ID))
}

func RBCReceive(ctx context.Context, p *party.HonestParty, sender uint32, ID []byte) *protobuf.Message {

	var EchoMessageMap = make(map[string]int)
	var ReadyMessageMap = make(map[string]int)

	var proposal *protobuf.Message
	var finishRBC = make(chan bool, 1)

	var isReadySent = false
	var mutex sync.Mutex
	var mutexEchoMap sync.Mutex
	var mutexReadyMap sync.Mutex

	var hasReceivedReady = make(map[uint32]bool)
	for i := uint32(0); i < p.N; i++ {
		hasReceivedReady[i] = false
	}

	// Handle RBCPropose message
	go func() {
		m := <-p.GetMessage("RBCPropose", ID)
		p.Log.Tracef("[party %v] receive RBCPropose message from [party %v], RBC instance ID: %s", p.PID, m.Sender, hex.EncodeToString(ID))
		proposal = m
		M := m.Data
		hLocal := sha256.New()
		hLocal.Write(M)

		EchoData, _ := proto.Marshal(&protobuf.RBCEcho{Hash: append(hLocal.Sum(nil))})
		p.Log.Tracef("[RBC][Party %v] broadcast RBCEcho message, instance ID: %s\n", p.PID, hex.EncodeToString(ID))
		err := p.Broadcast(&protobuf.Message{Type: "RBCEcho", Sender: p.PID, Id: ID, Data: EchoData})
		if err != nil {
			p.Log.Tracef("[party %v] multicast RBCEcho error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
		}
	}()

	// Handle RBCEcho message
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

				mutexEchoMap.Lock()
				counter, ok1 := EchoMessageMap[hash]
				if ok1 {
					EchoMessageMap[hash] = counter + 1
				} else {
					EchoMessageMap[hash] = 1
				}

				mutex.Lock()
				if !isReadySent && uint32(EchoMessageMap[hash]) >= p.N-p.F {
					isReadySent = true
					readyData, _ := proto.Marshal(&protobuf.RBCReady{Hash: []byte(hash)})
					err := p.Broadcast(&protobuf.Message{Type: "RBCReady", Sender: p.PID, Id: ID, Data: readyData})
					if err != nil {
						p.Log.Tracef("[party %v] multicast RBCReady message error: %v, RBC instance ID: %s", p.PID, err, hex.EncodeToString(ID))
					}
					p.Log.Tracef("[party %v] multicast RBCReady in the 2t+1 Echo case, RBC instance ID: %s", p.PID, hex.EncodeToString(ID))
				}
				mutex.Unlock()
				mutexEchoMap.Unlock()
				if isReadySent {
					break loop
				}
			}
		}
	}(ctx)

	// Handle RBCReady messages
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
				hash := string(payloadMessage.Hash)

				mutexReadyMap.Lock()
				counter, ok1 := ReadyMessageMap[hash]
				if ok1 {
					ReadyMessageMap[hash] = counter + 1
				} else {
					ReadyMessageMap[hash] = 1
				}

				mutex.Lock()
				if uint32(ReadyMessageMap[hash]) >= p.N-p.F {
					finishRBC <- true
					break
				}
				mutex.Unlock() //TODO: move this unlock into if statement
				mutexReadyMap.Unlock()
			}
		}
	}(ctx)

	select {
	case <-ctx.Done():
		return nil
	case <-finishRBC:
		p.Log.Tracef("[party %v] finish, RBC instance ID: %s\n", p.PID, hex.EncodeToString(ID))
		var replyMessage protobuf.Message
		proto.Unmarshal(proposal.Data, &replyMessage)
		return &replyMessage
		// return proposal
	}
}
