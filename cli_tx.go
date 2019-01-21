package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) createTx(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc, err := NewBlockchain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	UTXOSet := UTXOSet{bc.tip}

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	sendTx(localNodeAddress, tx)
}
