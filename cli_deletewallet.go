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

	fmt.Printf("Your address: %s is deleted\n", address)
}
