package dumbo_ng

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type store struct {
	data  [][]byte
	mutex *sync.Mutex
}

type lock struct {
	slot  uint32
	hash  []byte
	sig   []byte
	mutex *sync.Mutex
}

func (l *lock) set(e uint32, hash []byte, sig []byte) {
	l.mutex.Lock()
	if l.slot < e {
		l.slot = e
		l.hash = hash
		l.sig = sig
	}
	l.mutex.Unlock()
}

func (s *store) store(e uint32, data []byte) {
	s.mutex.Lock()
	if int(e) > len(s.data) {
		s.data = append(s.data, make([][]byte, int(e)-len(s.data))...)
	}
	s.data[e-1] = data
	s.mutex.Unlock()
}

func (s *store) load(e uint32) ([]byte, bool) {
	var v []byte
	var ok bool
	s.mutex.Lock()

	if e == 0 || int(e) > len(s.data) {
		ok = false
	} else if s.data[e-1] == nil {
		ok = false
	} else {
		ok = true
	}

	if ok {
		v = s.data[e-1]
	}
	s.mutex.Unlock()
	return v, ok
}

type TxWithTime struct {
	Txs  []byte
	Time time.Time
}

func EncodeTxWithTimeToBytes(txs []byte, time time.Time) []byte {
	txWithTime := TxWithTime{
		Txs:  txs,
		Time: time,
	}
	txWithTimeBytes, _ := json.Marshal(txWithTime)
	return txWithTimeBytes
}

func DecodeBytesToTxWithTime(txWithTimeBytes []byte) *TxWithTime {
	txWithTime := new(TxWithTime)
	err := json.Unmarshal(txWithTimeBytes, txWithTime)
	if err != nil {
		fmt.Println("Decode bytes to txWithTime error: ", err)
		return nil
	}
	return txWithTime
}
