package types

import (
	"container/list"
	"fmt"
	"sync"
)

type WrappedRawTx struct {
	tx       RawTransaction
	proposed bool
}

type MemPool struct {
	lock  *sync.Mutex
	order *list.List
	set   map[TxHash]*list.Element
}

func NewMemPool() *MemPool {
	c := &MemPool{
		set:   make(map[TxHash]*list.Element),
		lock:  new(sync.Mutex),
		order: new(list.List),
	}
	c.order.Init()
	return c
}

// showStatistic shows the statistic of the mempool
func (c *MemPool) showStatistic() {
	p := c.order.Front()
	marked := 0
	unmarked := 0
	for p != nil {
		if p.Value.(*WrappedRawTx).proposed {
			marked++
		} else {
			unmarked++
		}
		p = p.Next()
	}
	fmt.Println("marked: ", marked, "unmarked: ", unmarked, "total: ", marked+unmarked)
}

// Has return true if the given tx is in the mempool
func (c *MemPool) Has(tx RawTransaction) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.set[tx.Hash()]
	return ok
}

// Add add txs to the list and set, duplicate one will be ignored
func (c *MemPool) Add(txs ...RawTransaction) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, tx := range txs {
		txHash := tx.Hash()
		// avoid duplication
		if _, ok := c.set[txHash]; ok {
			continue
		}
		e := c.order.PushBack(&WrappedRawTx{
			tx:       tx,
			proposed: false,
		})
		c.set[txHash] = e
	}
}

// Remove remove commands from set and list
func (c *MemPool) Remove(txs []RawTransaction) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, tx := range txs {
		txHash := tx.Hash()
		if e, ok := c.set[txHash]; ok {
			c.order.Remove(e)
			delete(c.set, txHash)
		}
	}
}

// GetFirst return the top n unused commands from the list
func (c *MemPool) GetFirst(n int) []RawTransaction {
	c.lock.Lock()
	defer c.lock.Unlock()
	// defer c.showStatistic()

	if len(c.set) == 0 {
		return nil
	}
	txs := make([]RawTransaction, 0, n)
	i := 0
	// get the first element of list
	e := c.order.Front()
	for i < n {
		if e == nil {
			break
		}
		if wrtx := e.Value.(*WrappedRawTx); !wrtx.proposed {
			txs = append(txs, wrtx.tx)
			i++
		}
		e = e.Next()
	}
	return txs
}

func (c *MemPool) IsProposed(tx RawTransaction) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if e, ok := c.set[tx.Hash()]; ok {
		return e.Value.(*WrappedRawTx).proposed
	}
	return false
}

// MarkProposed will mark the given commands as proposed and move them to the back of the queue
func (c *MemPool) MarkProposed(txs []RawTransaction) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, tx := range txs {
		txHash := tx.Hash()
		if e, ok := c.set[txHash]; ok {
			e.Value.(*WrappedRawTx).proposed = true
			// Move to back so that it's not immediately deleted by a call to TrimToLen()
			c.order.MoveToBack(e)
		} else {
			// new tx, store it to back
			e := c.order.PushBack(&WrappedRawTx{tx: tx, proposed: true})
			c.set[txHash] = e
		}
	}
}

func (c *MemPool) UnMark(txs []RawTransaction) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, tx := range txs {
		if e, ok := c.set[tx.Hash()]; ok {
			e.Value.(*WrappedRawTx).proposed = false
			c.order.MoveToFront(e)
		}
	}
}
