package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// CreateOrLoadBlockchaindb ...
func CreateOrLoadBlockchaindb() {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
}
