package main

import (
	"fmt"
	"log"
)

func (cli *CLI) deleteWallet(address string) {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	delete(wallets.Wallets, address)
	wallets.SaveToFile()

	nodesInfo := NewNodesInfo()
	updateNodesInfo := make(map[string]*NodeInfo)

	updateNodesInfo[localNodeAddress] = nodesInfo[localNodeAddress]

	delete(updateNodesInfo[localNodeAddress].Wallets, address)

	UpdateNodesInfo(updateNodesInfo)

	fmt.Printf("Your address: %s is deleted\n", address)
}
