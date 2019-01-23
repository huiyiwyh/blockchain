package main

import (
	"log"
	"os"
)

func (cli *CLI) createTx(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Println("ERROR: Sender address is not valid")
		os.Exit(1)
	}
	if !ValidateAddress(to) {
		log.Println("ERROR: Recipient address is not valid")
		os.Exit(1)
	}

	bc, err := NewBlockchain()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	UTXOSet := UTXOSet{bc.tip}

	wallets, err := NewWallets()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	sendTx(localNodeAddress, tx)
}
