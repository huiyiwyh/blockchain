package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) getWalletsInfo() {
	wallets, err := NewWallets()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) deleteWallet(address string) {
	wallets, err := NewWallets()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	delete(wallets.Wallets, address)

	err = wallets.SaveToFile()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Your address: %s is deleted\n", address)
}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	err := wallets.SaveToFile()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Your new address: %s\n", address)
}
