package main

import (
	"fmt"
	"strconv"
)

func (cli *CLI) printBlockChain() {
	bc := NewBlockchain()
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
