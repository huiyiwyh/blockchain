package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

var blockChan chan *Block = make(chan *Block, 20)
var txChan chan *Transaction = make(chan *Transaction, 20)

// Mempool stores txs and processes txs
type Mempool struct {
	mempool map[string]*Transaction
	txNums  int
	mtx     *sync.Mutex
}

// AddTxs add txs into mempool
func (mp *Mempool) AddTxs(tx *Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNums++
	}

	fmt.Println("a new tx has been send to mempool")
}

// DeleteTxs delete txs from mempool
func (mp *Mempool) DeleteTxs(txs []*Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		if mp.mempool[txID] != nil {
			delete(mp.mempool, txID)
			mp.txNums--
		}
		fmt.Println("a tx was deleted from mempool")
	}
}

// GetTxNums returns the number of txs in the mempool
func (mp *Mempool) GetTxNums() int {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	txNums := mp.txNums
	return txNums
}

// VerifyTxs returns Transactions which is verified
func (mp *Mempool) VerifyTxs(bc *Blockchain) []*Transaction {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	var txs []*Transaction
	for id := range mp.mempool {
		tx := mp.mempool[id]
		if bc.VerifyTransaction(tx) {
			txs = append(txs, tx)
		} else {
			delete(mp.mempool, id)
			mp.txNums--
		}
	}

	return txs
}
