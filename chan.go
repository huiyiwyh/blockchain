package main

// BCMToS
// BCMToSSendBCM defines a chan that BlockChainManager sends BlockChainInfo to Server
var BCMToSSendBCM chan *BlockChainManagerInfo = make(chan *BlockChainManagerInfo, 20)

// BCMToSBlocksHash defines a chan that BlockChainManager sends blockshash to Server
var BCMToSBlocksHash chan *BlocksHash = make(chan *BlocksHash, 20)

// BCMToSBlockByHash defines a chan that BlockChainManager sends Block to server
var BCMToSBlockByHash chan *BlockByHash = make(chan *BlockByHash, 20)

// BCMToMTxs defines a chan that BlockChainManager sends txs to MempoolManager
var BCMToMTxs chan []*Transaction = make(chan []*Transaction, 20)

// MToBCMTxs defines a chan that MempoolManager sends txs to BlockChainManager
var MToBCMTxs chan []*Transaction = make(chan []*Transaction, 20)

// MToSTxs defines a chan that MempoolManager sends txs to Server
var MToSTxs chan []*Transaction = make(chan []*Transaction, 20)

// MToSSendM defines a chan that MempoolManager sends MempoolInfo to Server
var MToSSendM chan *MempoolManager = make(chan *MempoolManager, 20)

// MToSSendTxByHash defines a chan that MempoolManager sends tx to Server
var MToSSendTxByHash chan *TxByHash = make(chan *TxByHash, 20)

// SToBCMBlock defines a chan that Server sends Block to BlockChainManager
var SToBCMBlock chan *Block = make(chan *Block, 20)

// SToBCMGetBlockByHash defines a chan that Server sends GetBlockByHash to BlockChainManager
var SToBCMGetBlockByHash chan *BlockByHash = make(chan *BlockByHash, 20)

// SToBCMGetBlocksHash defines a chan that Server sends GetBlocksHash to BlockChainManager
var SToBCMGetBlocksHash chan *BlocksHash = make(chan *BlocksHash, 20)

// SToBCMGetBCM defines a chan that Server sends GetBlockChainManagerInfo to BlockChainManager
var SToBCMGetBCM chan *Notification = make(chan *Notification, 20)

// SToMTx defines a chan that Server sends Tx to MempoolManager
var SToMTx chan *Transaction = make(chan *Transaction, 20)

// SToMGetM defines a chan that Server sends GetMempoolInfo to MempoolManager
var SToMGetM chan *Notification = make(chan *Notification, 20)

// SToMGetTxByHash defines a chan that Server sends GetTxByHash to MempoolManager
var SToMGetTxByHash chan *TxByHash = make(chan *TxByHash, 20)

// Notification ...
type Notification struct{}

// BlocksHash ...
type BlocksHash struct {
	NodeFrom string
	Hashs    [][]byte
}

// BlockByHash ...
type BlockByHash struct {
	NodeFrom string
	Hash     []byte
	Block    *Block
}

// TxByHash ...
type TxByHash struct {
	NodeFrom string
	TxId     string
	Tx       *Transaction
}
