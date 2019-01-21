package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := CreateBlockchain(address)

	UTXOSet := UTXOSet{bc.tip}
	UTXOSet.Reindex()
}

func (cli *CLI) getblockchaininfo() {
	bc, err := NewBlockchain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bci := bc.Iterator(bc.tip)

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.BlockHeader.PrevBlockHash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if block.BlockHeader.PrevBlockHash == nil {
			break
		}
	}
}
