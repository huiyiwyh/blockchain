package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

// BlockVersion ...
const BlockVersion int = 1

// Block represents a block in the blockchain
type Block struct {
	BlockHeader    *BlockHeader
	Hash           []byte
	TransactionNum int
	Transactions   []*Transaction
	Height         int64
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
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int64) *Block {
	blockHeader := &BlockHeader{BlockVersion, prevBlockHash, []byte{}, time.Now().Unix(), 18, 0}
	block := &Block{blockHeader, []byte{}, len(transactions), transactions, height}
	block.BlockHeader.MerkleRootHash = block.HashTransactions()
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

// GetTracsactions returns the block's transactions
func (b *Block) GetTracsactions() []*Transaction {
	return b.Transactions
}

// GetTracsactionNum returns the block's tracsactionNum
func (b *Block) GetTracsactionNum() int {
	return b.TransactionNum
}

// GetHeight returns the block's height
func (b *Block) GetHeight() int64 {
	return b.Height
}

// GetHash returns the block's version
func (b *Block) GetHash() []byte {
	return b.Hash
}

// GetVersion returns the block's version
func (b *Block) GetVersion() int {
	return b.BlockHeader.Version
}

// GetTimestamp returns the block's Timestamp
func (b *Block) GetTimestamp() int64 {
	return b.BlockHeader.Timestamp
}

// GetPrevBlockHash returns the block's PrevBlockHash
func (b *Block) GetPrevBlockHash() []byte {
	return b.BlockHeader.PrevBlockHash
}

// GetNonce returns the block's Nonce
func (b *Block) GetNonce() int {
	return b.BlockHeader.Nonce
}

// GetMerkleRootHash returns the block's MerkleRootHash
func (b *Block) GetMerkleRootHash() []byte {
	return b.BlockHeader.MerkleRootHash
}

// GetBits returns the block's Bits
func (b *Block) GetBits() int64 {
	return b.BlockHeader.Bits
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
		//log.Panic(err)
	}

	return &block
}

// VerifyPoW ...
func (block *Block) VerifyPoW() bool {
	pow := NewProofOfWork(block)
	fmt.Println(pow.block.BlockHeader)
	if !pow.Validate() {
		return false
	}
	return true
}
