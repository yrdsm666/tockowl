package dumbo_ng

import (
	"bytes"
	"context"
	"log"
	"sync"
	"time"

	"github.com/yrdsm666/tockowl/consensus/bkr"
	"github.com/yrdsm666/tockowl/consensus/parbft/parbft1"
	"github.com/yrdsm666/tockowl/consensus/parbft/parbft2"
	"github.com/yrdsm666/tockowl/consensus/pkg/core"
	"github.com/yrdsm666/tockowl/consensus/pkg/party"
	"github.com/yrdsm666/tockowl/consensus/pkg/protobuf"
	"github.com/yrdsm666/tockowl/consensus/pkg/reedsolomon"
	"github.com/yrdsm666/tockowl/consensus/pkg/utils"
	"github.com/yrdsm666/tockowl/consensus/pkg/vectorcommitment"
	"github.com/yrdsm666/tockowl/consensus/sdumbo/smvba"
	"github.com/yrdsm666/tockowl/consensus/tockcat"
	"github.com/yrdsm666/tockowl/consensus/tockowl/mvba"
	"github.com/yrdsm666/tockowl/crypto"

	"github.com/shirou/gopsutil/mem"
	"github.com/vivint/infectious"
	"golang.org/x/crypto/sha3"
	"google.golang.org/protobuf/proto"
)

// MAXMESSAGE is the size of channels
var MAXMESSAGE = 1024

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func MainProgress(ctx context.Context, p *party.HonestParty, outputChannel chan [][]byte) {

	//store proposals
	var pStore = make([]*store, p.N)
	var pLock = make([]*lock, p.N)
	for i := uint32(0); i < p.N; i++ {
		pStore[i] = &store{
			data:  [][]byte{},
			mutex: new(sync.Mutex)}
		pLock[i] = &lock{
			slot:  0,
			hash:  []byte{},
			sig:   []byte{},
			mutex: new(sync.Mutex)}
	}
	var pCommit = make([]uint32, p.N)

	syncChannel := make([]chan uint32, p.N)
	for i := uint32(0); i < p.N; i++ {
		syncChannel[i] = make(chan uint32, 1024)
	}

	// pipeline proposer
	go proposer(ctx, p)
	//listen to proposers
	for i := uint32(0); i < p.N; i++ {
		go listener(ctx, p, i, pStore[i], pLock[i], syncChannel[i])
	}
	go Helper(p, pStore, pLock)

	if p.IsCrashNode(p.ID) {
		return
	}

	for e := uint32(1); ; e++ {
		//wait for MVBA's output

		//wait to invoke MVBA
		pids := []uint32{}
		slots := []uint32{}
		hashes := [][]byte{}
		sigs := [][]byte{}

		flags := make([]bool, p.N)
		count := uint32(0)
		for i := uint32(0); count < 2*p.F+1; i = (i + 1) % p.N {
			if !flags[i] {
				pLock[i].mutex.Lock()
				if pLock[i].slot > pCommit[i] {
					count++
					flags[i] = true
				}
				pLock[i].mutex.Unlock()
			}
		}

		for i := uint32(0); i < p.N; i++ {
			pLock[i].mutex.Lock()
			pids = append(pids, i)
			slots = append(slots, pLock[i].slot)
			hashes = append(hashes, pLock[i].hash)
			sigs = append(sigs, pLock[i].sig)
			pLock[i].mutex.Unlock()
		}

		value, err1 := proto.Marshal(&protobuf.BLockSetValue{
			Pid:  pids,
			Slot: slots,
			Hash: hashes,
		})
		validation, err2 := proto.Marshal(&protobuf.BLockSetValidation{
			Sig: sigs,
		})
		if err1 != nil || err2 != nil {
			log.Fatalln(err1, err2)
		}
		//wait for MVBA's output
		var resultValue []byte
		var resultValue2 [][]byte
		ctx, cancel := context.WithCancel(context.Background())

		switch p.Config.DumboNg.ConsensusType {
		case "tockcat": // TockCat
			resultValue = tockcat.MainProcess(ctx, p, utils.Uint32ToBytes(e), value, validation, Q)
		case "tockowl": // TockOwl
			resultValue = mvba.MainProcess(ctx, p, utils.Uint32ToBytes(e), value, validation, Q)
		case "tockowl+": // TockOwl+
			resultValue = mvba.MainProcess(ctx, p, utils.Uint32ToBytes(e), value, validation, Q)
		case "sdumbo": // sMVBA
			resultValue = smvba.MainProcess(ctx, p, utils.Uint32ToBytes(e), value, validation, Q)
		case "bkr": // BKR
			resultValue2 = bkr.MainProcess(p, utils.Uint32ToBytes(e), value)
		case "parbft1": // Parbft1
			resultValue = parbft1.MainProcess(p, utils.Uint32ToBytes(e), p.GetLeaderByEpoch(e), value, validation)
		case "parbft2": // Parbft2
			resultValue = parbft2.MainProcess(p, utils.Uint32ToBytes(e), p.GetLeaderByEpoch(e), value, validation)
		default:
			p.Log.Warn("[DUMBO-NG] Unsupported consensus type: ", p.Config.DumboNg.ConsensusType)
			return
		}
		cancel()

		if resultValue != nil {
			// The consensus component returns a single byte array, i.e. []byte
			var L protobuf.BLockSetValue
			var S protobuf.BLockSetValidation

			proto.Unmarshal(resultValue, &L)
			proto.Unmarshal(validation, &S)

			for i := uint32(0); i < p.N; i++ {
				_, ok := pStore[L.Pid[i]].load(L.Slot[i])
				if ok { // If ok is true, then the current slot is must >= L.slot[i] and locked slot >= L.slot[i]-1
					pLock[L.Pid[i]].set(L.Slot[i], L.Hash[i], S.Sig[i])
				}
			}

			output := obtainProposals(p, e, pStore, pLock, pCommit, L.Slot, L.Hash, S.Sig, syncChannel)
			outputChannel <- output
		} else if resultValue2 != nil {
			maxSlot := make([]uint32, p.N)
			maxHash := make([][]byte, p.N)
			maxSig := make([][]byte, p.N)

			// The consensus component returns n byte arrays, i.e. [][]byte
			for i := 0; i < len(resultValue2); i++ {
				if len(resultValue2[i]) == 0 {
					continue
				}
				var L protobuf.BLockSetValue //L={(j,s,h)}
				var S protobuf.BLockSetValidation

				proto.Unmarshal(resultValue2[i], &L)
				proto.Unmarshal(validation, &S)

				for i := uint32(0); i < p.N; i++ {
					_, ok := pStore[L.Pid[i]].load(L.Slot[i])
					if ok { // If ok is true, then the current slot is must >= L.slot[i] and locked slot >= L.slot[i]-1
						pLock[L.Pid[i]].set(L.Slot[i], L.Hash[i], S.Sig[i])
					}

					if L.Slot[i] > maxSlot[i] {
						maxSlot[i] = L.Slot[i]
						maxHash[i] = L.Hash[i]
						maxSig[i] = S.Sig[i]
					}
				}

			}
			output := obtainProposals(p, e, pStore, pLock, pCommit, maxSlot, maxHash, maxSig, syncChannel)
			outputChannel <- output
		} else {
			p.Log.Warn("[DUMBO-NG] Consensus failed: ", p.Config.DumboNg.ConsensusType)
			return
		}

	}

}

func proposer(ctx context.Context, p *party.HonestParty) {
	var hash, signature []byte

	rawtxs := utils.GetTxs(p.Config.BatchSize, p.Config.PayloadSize)

	for s := uint32(1); ; s++ { //slot
		select {
		case <-ctx.Done():
			return
		default:
			var buf1 bytes.Buffer
			buf1.Write(utils.Uint32ToBytes(p.PID))
			buf1.Write(utils.Uint32ToBytes(s))
			ID := buf1.Bytes()

			startTime := time.Now()
			tx := EncodeTxWithTimeToBytes(rawtxs, startTime)

			var proposalMessage *protobuf.Message
			if s == 1 {
				proposalMessage = core.Encapsulation("NGProposal", ID, p.PID, &protobuf.NGProposal{
					Tx: tx,
				})
			} else {
				proposalMessage = core.Encapsulation("NGProposal", ID, p.PID, &protobuf.NGProposal{
					Tx:   tx,
					Hash: hash,      //hash of previous slot
					Sig:  signature, // sig on previous slot
				})
			}

			p.Broadcast(proposalMessage)

			// var buf2 bytes.Buffer
			sigs := [][]byte{}
			ids := []int64{}
			h := sha3.Sum512(tx)
			// buf2.Write([]byte("Proposal"))
			// buf2.Write(ID)
			// buf2.Write(h[:])
			// sm := buf2.Bytes()

			for {
				m := <-p.GetMessage("Received", ID)
				payload := core.Decapsulation("Received", m).(*protobuf.Received)

				sigs = append(sigs, payload.Sigshare)
				ids = append(ids, int64(m.Sender)+1)
				if len(sigs) > int(2*p.F) {
					hash = h[:]
					signature, _ = crypto.CombineSignatures(sigs, ids)
					break
				}
			}
		}
	}
}

func listener(ctx context.Context, p *party.HonestParty, j uint32, pStore *store, pLock *lock, syncChannel chan uint32) {
	var preID, ID []byte

	for s := uint32(1); ; s++ {
		preID = ID
		var buf1 bytes.Buffer
		buf1.Write(utils.Uint32ToBytes(j))
		buf1.Write(utils.Uint32ToBytes(s))
		ID = buf1.Bytes()
	slotFlag:
		for {
			select {
			case <-ctx.Done():
				return
			case commitedSlot := <-syncChannel:
				if commitedSlot >= s {
					s = commitedSlot
					var buf2 bytes.Buffer
					buf2.Write(utils.Uint32ToBytes(j))
					buf2.Write(utils.Uint32ToBytes(s))
					ID = buf2.Bytes()
					break slotFlag
				}
			case m := <-p.GetMessage("NGProposal", ID):
				payload := (core.Decapsulation("NGProposal", m)).(*protobuf.NGProposal)

				var buf2 bytes.Buffer
				h := sha3.Sum512(payload.Tx)
				buf2.Write([]byte("Proposal"))
				buf2.Write(ID)
				buf2.Write(h[:])
				sm := buf2.Bytes()

				if s == 1 {
					pStore.store(s, payload.Tx)
					sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("Echo"||ID||h)
					receivedMessage := core.Encapsulation("Received", ID, p.PID, &protobuf.Received{
						Sigshare: sigShare,
					})
					if j == p.PID {
						msgByte, _ := proto.Marshal(receivedMessage)
						p.MsgByteEntrance <- msgByte
					} else {
						p.Unicast(p.GetNetworkInfo()[int64(j)], receivedMessage)
					}
				} else {
					// preTx, _ := pStore.load(s - 1)

					var buf3 bytes.Buffer
					buf3.Write([]byte("Proposal"))
					buf3.Write(preID)
					buf3.Write(payload.Hash)
					presm := buf3.Bytes()
					err := crypto.VerifySignature(p.Config.Keys.VerifyKeys[int(m.Sender)], presm, payload.Sig)

					preTx, _ := pStore.load(s - 1)
					preHash := sha3.Sum512(preTx)
					if err == nil && bytes.Equal(payload.Hash, preHash[:]) {
						pLock.set(s-1, payload.Hash, payload.Sig)
						pStore.store(s, payload.Tx)
						sigShare := crypto.ThresholdSign(p.Config.Keys.SecretKey, sm) //sign("Echo"||ID||h)
						receivedMessage := core.Encapsulation("Received", ID, p.PID, &protobuf.Received{
							Sigshare: sigShare,
						})
						if j == p.PID {
							msgByte, _ := proto.Marshal(receivedMessage)
							p.MsgByteEntrance <- msgByte
						} else {
							p.Unicast(p.GetNetworkInfo()[int64(j)], receivedMessage)
						}
					}
				}
				break slotFlag
			}
		}

	}

}

func Helper(p *party.HonestParty, pStore []*store, pLock []*lock) {
	coder := reedsolomon.NewRScoder(int(p.F+1), int(p.N))
	for {
		m := <-p.GetMessage("NGCallHelp", []byte{})
		payload := core.Decapsulation("NGCallHelp", m).(*protobuf.NGCallHelp)

		pLock[payload.Pid].mutex.Lock()
		locked := pLock[payload.Pid].slot
		pLock[payload.Pid].mutex.Unlock()

		if locked >= payload.Slot {
			value, _ := pStore[payload.Pid].load(payload.Slot)

			temp := coder.Encode(value)
			fragments := make([][]byte, p.N)
			for i := uint32(0); i < p.N; i++ {
				fragments[i] = temp[i].Data
			}
			vCommiter, _ := vectorcommitment.NewMerkleTree(fragments)
			root := vCommiter.GetMerkleTreeRoot()
			proof1, proof2 := vCommiter.GetMerkleTreeProof(int(p.PID))

			var IDbuf bytes.Buffer
			IDbuf.Write(utils.Uint32ToBytes(payload.Pid))
			helpMessage := core.Encapsulation("NGHelp", IDbuf.Bytes(), p.PID, &protobuf.NGHelp{
				Pid:    payload.Pid,
				Slot:   payload.Slot,
				Shard:  fragments[p.PID],
				Root:   root,
				Proof1: proof1,
				Proof2: proof2,
			})
			if m.Sender == p.PID {
				msgByte, _ := proto.Marshal(helpMessage)
				p.MsgByteEntrance <- msgByte
			} else {
				p.Unicast(p.GetNetworkInfo()[int64(m.Sender)], helpMessage)
			}
		} else {
			value, ok1 := pStore[payload.Pid].load(payload.Slot)
			if ok1 {
				h := sha3.Sum512(value)
				var buf bytes.Buffer
				buf.Write([]byte("Proposal"))
				buf.Write(utils.Uint32ToBytes(payload.Pid))
				buf.Write(utils.Uint32ToBytes(payload.Slot))
				buf.Write(h[:])
				sm := buf.Bytes()
				err := crypto.VerifySignature(p.Config.Keys.VerifyKeys[int(m.Sender)], sm, payload.Sig)
				if err == nil {
					pLock[payload.Pid].set(payload.Slot, h[:], payload.Sig)

					temp := coder.Encode(value)
					fragments := make([][]byte, p.N)
					for i := uint32(0); i < p.N; i++ {
						fragments[i] = temp[i].Data
					}
					vCommiter, _ := vectorcommitment.NewMerkleTree(fragments)
					root := vCommiter.GetMerkleTreeRoot()
					proof1, proof2 := vCommiter.GetMerkleTreeProof(int(p.PID))

					var IDbuf bytes.Buffer
					IDbuf.Write(utils.Uint32ToBytes(payload.Pid))
					helpMessage := core.Encapsulation("NGHelp", IDbuf.Bytes(), p.PID, &protobuf.NGHelp{
						Pid:    payload.Pid,
						Slot:   payload.Slot,
						Shard:  fragments[p.PID],
						Root:   root,
						Proof1: proof1,
						Proof2: proof2,
					})
					if m.Sender == p.PID {
						msgByte, _ := proto.Marshal(helpMessage)
						p.MsgByteEntrance <- msgByte
					} else {
						p.Unicast(p.GetNetworkInfo()[int64(m.Sender)], helpMessage)
					}
				}
			}
		}

	}
}

func CallHelp(p *party.HonestParty, pStore []*store, pLock []*lock, j uint32, maxSlot uint32, maxHash []byte, maxSig []byte, wg *sync.WaitGroup) { //pid, slot, hash, sig, return channel

	pLock[j].mutex.Lock()
	locked := pLock[j].slot
	pLock[j].mutex.Unlock()

	var buf bytes.Buffer
	buf.Write([]byte("Proposal"))
	buf.Write(utils.Uint32ToBytes(j))
	buf.Write(utils.Uint32ToBytes(maxSlot))
	buf.Write(maxHash)
	sm := buf.Bytes()
	err := crypto.VerifySignature(p.Config.Keys.VerifyKeys[int(j)], sm, maxSig)

	coder := reedsolomon.NewRScoder(int(p.F+1), int(p.N))

	if maxSlot > locked && err == nil {

		shards := make([][]infectious.Share, maxSlot-locked)

		for k := locked + 1; k <= maxSlot; k++ {
			if k < maxSlot {
				callMessage := core.Encapsulation("CallHelp", []byte{}, p.PID, &protobuf.NGCallHelp{
					Pid:  j,
					Slot: k,
				})
				p.Broadcast(callMessage)
			} else {
				callMessage := core.Encapsulation("CallHelp", []byte{}, p.PID, &protobuf.NGCallHelp{
					Pid:  j,
					Slot: k,
					Sig:  maxSig,
				})
				p.Broadcast(callMessage)
			}
		}

		flag := make([]bool, maxSlot-locked)
		count := 0

		for {
			var IDbuf bytes.Buffer
			IDbuf.Write(utils.Uint32ToBytes(j))
			m := <-p.GetMessage("NGHelp", IDbuf.Bytes())
			payload := core.Decapsulation("NGHelp", m).(*protobuf.NGHelp)
			if payload.Slot <= locked || payload.Slot > maxSlot || flag[payload.Slot-locked-1] { //drop old mesages
				continue
			}
			if vectorcommitment.VerifyMerkleTreeProof(payload.Root, payload.Proof1, payload.Proof2, payload.Shard) {
				shards[payload.Slot-locked-1] = append(shards[payload.Slot-locked-1], infectious.Share{
					Data:   payload.Shard,
					Number: int(m.Sender),
				})
			}
			if len(shards[payload.Slot-locked-1]) == int(p.F+1) {
				value, err := coder.Decode(shards[payload.Slot-locked-1]) //decode
				if err != nil {
					panic(err)
				}
				pStore[j].store(payload.Slot, value)
				flag[payload.Slot-locked-1] = true
				count++
				if count == int(maxSlot-locked) {
					break
				}
			}
		}
		pLock[j].set(maxSlot, maxHash, maxSig)

	}
	wg.Done()
}

func obtainProposals(p *party.HonestParty, e uint32, pStore []*store, pLock []*lock, pCommit []uint32, certSlot []uint32, certHash [][]byte, certSig [][]byte, syncChannel []chan uint32) [][]byte {

	var wg sync.WaitGroup

	for i := uint32(0); i < p.N; i++ {
		pLock[i].mutex.Lock()
		if pLock[i].slot < certSlot[i] { //CallHelp
			wg.Add(1)
			go CallHelp(p, pStore, pLock, i, certSlot[i], certHash[i], certSig[i], &wg)
		}
		pLock[i].mutex.Unlock()
	}

	wg.Wait()

	output := make([][]byte, 0)

	for i := uint32(0); i < p.N; i++ {
		for k := pCommit[i] + 1; k <= certSlot[i]; k++ {
			value, _ := pStore[i].load(k)
			output = append(output, value)
		}

		pLock[i].set(certSlot[i], certHash[i], certSig[i])

		pCommit[i] = certSlot[i]

		syncChannel[i] <- certSlot[i]
	}
	return output
}

func Q(p *party.HonestParty, ID []byte, value []byte, validation []byte, hashVerifyMap *sync.Map, sigVerifyMap *sync.Map) error {
	return nil
}
