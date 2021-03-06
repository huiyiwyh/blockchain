package main

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator is used to iterate over blockchain blocks
type BlockChainIterator struct {
	currentHash []byte
}

// Iterator returns BlockChainIterator from the lastHash
func (b *Blockchain) Iterator(lastHash []byte) *BlockChainIterator {
	bc := &BlockChainIterator{lastHash}
	return bc
}

// Next returns next block starting from the tip
func (i *BlockChainIterator) Next() *Block {
	var block *Block

	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockdata := b.Get(i.currentHash)
		block = DeserializeBlock(blockdata)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.BlockHeader.PrevBlockHash

	return block
}
