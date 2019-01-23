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
type MempoolManagerInfo struct {
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
		case <-SToMGetM:
			go mp.returnServerMempoolManagerInfo()
		case txByHash := <-SToMGetTxByHash:
			go mp.GetTx(txByHash)
		}
	}
}

func (mp *MempoolManager) returnServerMempoolManagerInfo() {
	nmpm := mp.newMempoolManagerInfo()
	MToSSendMMI <- nmpm
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

			for _, tx := range mp.mempool {
				txs = append(txs, tx)
				delete(mp.mempool, hex.EncodeToString(tx.ID))
				mp.txNum--
			}

			MToBGetBCMI <- &Notification{}
			nbcm := <-BToMSendBCMI

			u := &UTXOSet{nbcm.Hash}

			for _, tx := range txs {
				if u.VerifyTransaction(tx) != true {
					log.Println("ERROR: Invalid transaction")
				}
			}
			MToBTxs <- txs
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
func (mp *MempoolManager) newMempoolManagerInfo() *MempoolManagerInfo {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()

	return &MempoolManagerInfo{mp.mempool, mp.txNum}
}
