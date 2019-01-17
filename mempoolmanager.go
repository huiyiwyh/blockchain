package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// MempoolManager ...
type MempoolManager struct {
	mempool map[string]*Transaction
	txNum   int
	mtx     *sync.Mutex
}

// NewMempoolManager returns a new MempoolManager
func NewMempoolManager() *MempoolManager {
	return &MempoolManager{make(map[string]*Transaction), 0, new(sync.Mutex)}
}

// MempoolManagerProcessor manages received txs
func (mp *MempoolManager) MempoolManagerProcessor() {
	for {
		select {
		case tx := <-SToMTx:
			go mp.ValidateTransactionIsValidAndAdd(tx)
		case txs := <-BCMToMTxs:
			go mp.DeleteTxs(txs)
		case <-SToMGetM:
			go mp.GetMempoolManager()
		case txByHash := <-SToMGetTxByHash:
			go mp.GetTx(txByHash)
		}
	}
}

// ValidateTransactionIsValidAndAdd ...
func (mp *MempoolManager) ValidateTransactionIsValidAndAdd(tx *Transaction) {
	mp.AddTxs(tx)
	fmt.Println("txs in mempool's number :", mp.GetTxNum())

	mp.MaybeSendTxsToBCM()
}

// MaybeSendTxsToBCM ...
func (mp *MempoolManager) MaybeSendTxsToBCM() {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	txNum := mp.txNum
	if txNum > 1 {
		var txs []*Transaction

		for _, tx := range mp.mempool {
			txs = append(txs, tx)
		}
		MToBCMTxs <- txs
	}
}

// AddTxs add txs into mempool
func (mp *MempoolManager) AddTxs(tx *Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNum++
		fmt.Println("a new tx has been send to mempool")
		return
	}

	fmt.Println("the new tx has already in mempool")
}

// DeleteTxs delete txs from mempool
func (mp *MempoolManager) DeleteTxs(txs []*Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		if mp.mempool[txID] != nil {
			delete(mp.mempool, txID)
			mp.txNum--
		}
		fmt.Println("a tx was deleted from mempool")
	}
}

// GetTx ...
func (mp *MempoolManager) GetTx(txByHash *TxByHash) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[txByHash.TxId] != nil {
		txByHash.Tx = mp.mempool[txByHash.TxId]
		MToSSendTxByHash <- txByHash
	}
}

// GetTxs ...
func (mp *MempoolManager) GetTxs() []*Transaction {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	var txs []*Transaction
	for _, tx := range mp.mempool {
		txs = append(txs, tx)
	}

	return txs
}

// GetTxNum ...
func (mp *MempoolManager) GetTxNum() int {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	txNum := mp.txNum
	return txNum
}

// GetMempoolManager ...
func (mp *MempoolManager) GetMempoolManager() {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	nmp := &MempoolManager{mp.mempool, mp.txNum, nil}

	MToSSendM <- nmp
}
