package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// CreateOrLoadBlockChaindb ...
func CreateOrLoadBlockChaindb() {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
}
