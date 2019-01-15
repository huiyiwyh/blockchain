package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// BlockVersion ...
const BlockVersion int = 1

// Block represents a block in the blockchain
type Block struct {
	BlockHeader     *BlockHeader
	Hash            []byte
	TransactionsNum int
	Transactions    []*Transaction
	Height          int
}

// BlockHeader represents a blockheader in the block
type BlockHeader struct {
	Version        int
	PrevBlockHash  []byte
	MerkleRootHash []byte
	Timestamp      int64
	Bits           int64
	Nonce          int
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	blockHeader := &BlockHeader{BlockVersion, prevBlockHash, []byte{}, time.Now().Unix(), 19, 0}
	block := &Block{blockHeader, []byte{}, len(transactions), transactions, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	if nonce == -1 && hash == nil {
		return nil
	}

	block.Hash = hash[:]
	block.BlockHeader.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, nil, 0)
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// VerifyPoW ...
func (block *Block) VerifyPoW() bool {
	pow := NewProofOfWork(block)
	if !pow.Validate() {
		return false
	}
	return true
}
