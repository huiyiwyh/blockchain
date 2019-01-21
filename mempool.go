package main

import (
	"encoding/hex"
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

// Processor manages received txs
func (mp *MempoolManager) Processor() {
	for {
		select {
		case tx := <-SToMTx:
			go mp.MaybeSendTxsToBCM(tx)
		case txs := <-BCMToMTxs:
			go mp.DeleteTxs(txs)
		case <-SToMGetM:
			go mp.GetMempoolManagerInfo()
		case txByHash := <-SToMGetTxByHash:
			go mp.GetTx(txByHash)
		}
	}
}

// MaybeSendTxsToBCM ...
func (mp *MempoolManager) MaybeSendTxsToBCM(tx *Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNum++
		//fmt.Println("a new tx has been send to mempool")
		//fmt.Println("txs in mempool's number :", mp.txNum)
	}

	txNum := mp.txNum
	if txNum > 0 {
		var txs []*Transaction

		for _, tx := range mp.mempool {
			txs = append(txs, tx)
			delete(mp.mempool, hex.EncodeToString(tx.ID))
			mp.txNum--
		}
		MToBCMTxs <- txs
	}
}

// AddTx ...
func (mp *MempoolManager) AddTx(tx *Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNum++
		//fmt.Println("a new tx has been send to mempool")
		//fmt.Println("txs in mempool's number :", mp.txNum)
	}
}

// DeleteTxs ...
func (mp *MempoolManager) DeleteTxs(txs []*Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		if mp.mempool[txID] != nil {
			delete(mp.mempool, txID)
			mp.txNum--
			//fmt.Println("a tx was deleted from mempool")
		}
	}
}

// GetTx ...
func (mp *MempoolManager) GetTx(txByHash *TxByHash) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	txByHash.Tx = mp.mempool[txByHash.TxId]
	MToSSendTxByHash <- txByHash
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

// GetMempoolManagerInfo ...
func (mp *MempoolManager) GetMempoolManagerInfo() {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	nmp := &MempoolManager{mp.mempool, mp.txNum, nil}

	MToSSendM <- nmp
}
