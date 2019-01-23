package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
)

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
}

// CreateBlockchain creates a new Blockchain DB
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
			return fmt.Errorf("The Blockchain is exists")
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

// LoadTopBlock ...
func LoadTopBlock() (*Block, error) {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var block *Block

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			fmt.Println("the Blockchain is not exist")
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		block = lastBlock

		return nil
	})
	if err != nil {
		panic(err)
	}

	return block, nil
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() (*Blockchain, error) {
	resp, err := http.Get("http://127.0.0.1:8080/getBlockchainMangerInfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var blockchainInfo BlockchainInfo

	dec := gob.NewDecoder(resp.Body)
	err = dec.Decode(&blockchainInfo)

	bc := &Blockchain{blockchainInfo.Hash}
	return bc, nil
}
