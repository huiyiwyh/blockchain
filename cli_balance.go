package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc, err := NewBlockchain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	UTXOSet := UTXOSet{bc.tip}

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXOByHash(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
