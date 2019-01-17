package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(address string) *Blockchain {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var tip []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b != nil {
			return fmt.Errorf("The blockchain is exists")
		}

		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		_, err = tx.CreateBucket([]byte(orphanBlocksBucket))
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip}

	return &bc
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip}

	return &bc
}
