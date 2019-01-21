package main

import (
	"fmt"
	"log"
)

func (cli *CLI) getWalletsInfo() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) deleteWallet(address string) {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	delete(wallets.Wallets, address)
	wallets.SaveToFile()

	fmt.Printf("Your address: %s is deleted\n", address)
}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}
