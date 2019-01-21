package main

type msgSendTx struct {
	From   []byte
	To     []byte
	Amount int
}

type msgGetTxByHash struct {
	TxHash []byte
}

type msgGetBlockchainInfo struct {
	Address string
}

type msgGetBalance struct {
	Address string
}

type msgGetWalletsInfo struct {
}

type msgGetBCM struct {
}
