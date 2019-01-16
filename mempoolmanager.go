package main

import "fmt"

var txChanFromMempoolToBlockChainManager chan []*Transaction = make(chan []*Transaction, 20)
var txChanFromMempoolToServer chan []*Transaction = make(chan []*Transaction, 20)

// MempoolManager manages received txs
func MempoolManager() {
	select {
	case tx := <-txChanFromServerToMempool:
		go ValidateTransactionIsValidAndAdd(tx)
	case block := <-blockChanFromBlockChainManagerToMempool:
		go DeleteTxsHasBeenMindIntoBlockInMempool(block)
	}
}

// ValidateTransactionIsValidAndAdd ...
func ValidateTransactionIsValidAndAdd(tx *Transaction) {
	mp.AddTxs(tx)

	fmt.Println("txs in mempool's number :", mp.GetTxNums())

	// if mp.GetTxsNumAndRetrieve() >  {
	// 	bc := NewBlockchain()
	// 	txs := mp.VerifyTxs(bc)

	// 	if len(txs) == 0 {
	// 		fmt.Printf("All transactions are added into block! Waiting for new ones...\n")
	// 		return
	// 	}
	// }
}

// DeleteTxsHasBeenMindIntoBlockInMempool delete txs in mempool
func DeleteTxsHasBeenMindIntoBlockInMempool(block *Block) {
	txs := block.GetTracsactions()
	mp.DeleteTxs(txs)
}
