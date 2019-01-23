package main

// BCMToCSendBCM defines a chan that BlockchainManager sends BlockchainManagerInfo to Client
var BToCBCMI chan *BlockchainManagerInfo = make(chan *BlockchainManagerInfo, 20)

// BCMToSSendBCM defines a chan that BlockchainManager sends BlockchainInfo to Server
var BToSBCMI chan *BlockchainManagerInfo = make(chan *BlockchainManagerInfo, 20)

// BCMToSBlocksHash defines a chan that BlockchainManager sends blockshash to Server
var BToSBlocksHash chan *BlocksHash = make(chan *BlocksHash, 20)

// BCMToSBlockByHash defines a chan that BlockchainManager sends Block to server
var BToSBlockByHash chan *BlockByHash = make(chan *BlockByHash, 20)

// BCMToMTxs defines a chan that BlockchainManager sends txs to MempoolManager
var BToMTxs chan []*Transaction = make(chan []*Transaction, 20)

// BCMToMSendBCM defines a chan that BlockchainManager sends BlockchainManagerInfo to MempoolManager
var BToMBCMI chan *BlockchainManagerInfo = make(chan *BlockchainManagerInfo, 20)

// BToSBlock defines a chan that BlockchainManager sends block to Server
var BToSBlock chan *Block = make(chan *Block, 20)

// MToBCMGetBCM defines a chan that MempoolManager sends GetBlockchainManagerInfo to BlockchainManager
var MToBGetBCMI chan *Notification = make(chan *Notification, 20)

// MToBCMTxs defines a chan that MempoolManager sends txs to BlockchainManager
var MToBTxs chan []*Transaction = make(chan []*Transaction, 20)

// MToSTxs defines a chan that MempoolManager sends txs to Server
var MToSTxs chan []*Transaction = make(chan []*Transaction, 20)

// MToSSendM defines a chan that MempoolManager sends MempoolInfo to Server
var MToSMMI chan *MempoolManagerInfo = make(chan *MempoolManagerInfo, 20)

// MToSSendTxByHash defines a chan that MempoolManager sends tx to Server
var MToSSendTxByHash chan *TxByHash = make(chan *TxByHash, 20)

// SToBCMBlock defines a chan that Server sends Block to BlockchainManager
var SToBBlock chan *Block = make(chan *Block, 20)

// SToBCMGetBlockByHash defines a chan that Server sends GetBlockByHash to BlockchainManager
var SToBGetBlockByHash chan *BlockByHash = make(chan *BlockByHash, 20)

// SToBCMGetBlocksHash defines a chan that Server sends GetBlocksHash to BlockchainManager
var SToBGetBlocksHash chan *BlocksHash = make(chan *BlocksHash, 20)

// SToBCMGetBCM defines a chan that Server sends GetBlockchainManagerInfo to BlockchainManager
var SToBGetBCMI chan *Notification = make(chan *Notification, 20)

// SToMTx defines a chan that Server sends Tx to MempoolManager
var SToMTx chan *Transaction = make(chan *Transaction, 20)

// SToMGetM defines a chan that Server sends GetMempoolManagerInfo to MempoolManager
var SToMGetMMI chan *Notification = make(chan *Notification, 20)

// SToMGetTxByHash defines a chan that Server sends GetTxByHash to MempoolManager
var SToMGetTxByHash chan *TxByHash = make(chan *TxByHash, 20)

// SToPPeer defines a chan that Server sends peers to PeerManager
var SToPPeer chan []string = make(chan []string, 20)

// SToPGetPM defines a chan that Server sends GetPoolManagerInfo to PeerManager
var SToPGetPMI chan *Notification = make(chan *Notification, 20)

// CToBCMGetBCMI defines a chan that Client sends GetBlockchanManagerInfo to BlockchainManagerInfo
var CToBGetBCMI chan *Notification = make(chan *Notification, 20)

// PToSSendPM defines a chan that PeerManager sends PeerManagerInfo to Server
var PToSPMI chan *PeerManagerInfo = make(chan *PeerManagerInfo, 20)

// ReceivedBlock defines a chan that when received a block will stop mining
var ReceivedBlock chan *Notification = make(chan *Notification, 20)

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
