package main

import "fmt"

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	nodesInfo := NewNodesInfo()

	updateNodesInfo := make(map[string]*NodeInfo)

	if nodesInfo[localNodeAddress] == nil {
		nodesInfo[localNodeAddress] = &NodeInfo{}
	}

	if nodesInfo[localNodeAddress].Wallets == nil || len(nodesInfo[localNodeAddress].Wallets) <= 0 {
		updateNodesInfo[localNodeAddress] = nodesInfo[localNodeAddress]
		updateNodesInfo[localNodeAddress].Wallets = make(map[string]string)
	}
	updateNodesInfo[localNodeAddress] = nodesInfo[localNodeAddress]
	updateNodesInfo[localNodeAddress].Wallets[address] = address

	UpdateNodesInfo(updateNodesInfo)

	fmt.Printf("Your new address: %s\n", address)
}
