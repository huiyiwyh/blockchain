package main

import (
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// BlockChain implements interactions with a DB
type BlockChain struct {
	tip []byte
}

// CreateBlockChain creates a new blockchain DB
func CreateBlockChain(address string) *BlockChain {
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

	bc := BlockChain{tip}

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
			fmt.Println("the blockchain is not exist")
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

// NewBlockChain creates a new Blockchain with genesis Block
func NewBlockChain() *BlockChain {
	CToBCMGetBCM <- &Notification{}

	nbcm := <-BCMToCSendBCM
	bc := &BlockChain{nbcm.Hash}

	return bc
}
