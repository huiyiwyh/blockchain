package main

import (
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := CreateBlockChain(address)

	UTXOSet := UTXOSet{bc.tip}
	UTXOSet.Reindex()
}
