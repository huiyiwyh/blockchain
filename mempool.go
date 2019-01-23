package main

import (
	"encoding/hex"
	"log"
	"sync"
	"time"
)

// MempoolManager ...
type MempoolManager struct {
	mempool map[string]*Transaction
	txNum   int
	mtx     *sync.Mutex
}

// MempoolManagerInfo ...
type MempoolInfo struct {
	mempool map[string]*Transaction
	txNum   int
}

// NewMempoolManager returns a new MempoolManager
func NewMempoolManager() *MempoolManager {
	return &MempoolManager{make(map[string]*Transaction), 0, new(sync.Mutex)}
}

// Processor manages received txs
func (mp *MempoolManager) Processor() {
	go mp.maybeSendTxsToBCM()

	for {
		select {
		case tx := <-SToMTx:
			go mp.AddTx(tx)
		case txs := <-BToMTxs:
			go mp.DeleteTxs(txs)
		case <-SToMGetMI:
			go mp.returnServerMempoolManagerInfo()
		case txByHash := <-SToMGetTxByHash:
			go mp.GetTx(txByHash)
		}
	}
}

func (mp *MempoolManager) returnServerMempoolManagerInfo() {
	mi := mp.newMempoolInfo()
	MToSMI <- mi
}

// MaybeSendTxsToBCM ...
func (mp *MempoolManager) maybeSendTxsToBCM() {
	for {
		time.Sleep(1 * time.Second)
		mp.mtx.Lock()
		if time.Now().Unix()%20 != 0 {
			mp.mtx.Unlock()
			continue
		}

		if mp.txNum > 0 {
			var txs []*Transaction

			MToBGetBI <- &Notification{}
			nbcm := <-BToMBI

			u := &UTXOSet{nbcm.Hash}

			for _, tx := range mp.mempool {
				if u.VerifyTransaction(tx) != true {
					log.Println("ERROR: Invalid transaction")
				} else {
					txs = append(txs, tx)
				}
				delete(mp.mempool, hex.EncodeToString(tx.ID))
				mp.txNum--
			}

			if len(txs) > 0 {
				MToBTxs <- txs
			}
		}
		mp.mtx.Unlock()
	}
}

// AddTx ...
func (mp *MempoolManager) AddTx(tx *Transaction) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNum++
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
		}
	}
}

// GetTx ...
func (mp *MempoolManager) GetTx(txByHash *TxByHash) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	txByHash.Tx = mp.mempool[txByHash.TxId]
	MToSTxByHash <- txByHash
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
func (mp *MempoolManager) newMempoolInfo() *MempoolInfo {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return &MempoolInfo{mp.mempool, mp.txNum}
}
