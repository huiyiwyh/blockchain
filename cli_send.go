package main

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain()
	UTXOSet := UTXOSet{bc}

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	nodesInfo := NewNodesInfo()
	for _, node := range nodesInfo[localNodeAddress].Nodes {
		if node != localNodeAddress {
			sendTx(node, tx)
			fmt.Printf("send to %s\n", node)
		}
	}

	fmt.Printf("Success!\n")
}
