package main

import (
	"fmt"
	"os"
)

func (cli *CLI) reindexUTXO() {
	bc, err := NewBlockchain()
	if err != nil {
		os.Exit(1)
	}

	UTXOSet := UTXOSet{bc.tip}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
