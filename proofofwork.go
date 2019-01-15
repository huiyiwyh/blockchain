package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// func GetMedianTime() {

// }

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// if GetMedianTime() ....
	target.Lsh(target, uint(256-b.BlockHeader.Bits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	fmt.Println("pow.block.BlockHeader.PrevBlockHash", pow.block.BlockHeader.PrevBlockHash)
	fmt.Println("pow.block.HashTransactions()", pow.block.HashTransactions())
	fmt.Println("IntToHex(pow.block.BlockHeader.Timestamp)", IntToHex(pow.block.BlockHeader.Timestamp))
	fmt.Println("IntToHex(int64(pow.block.BlockHeader.Bits))", IntToHex(int64(pow.block.BlockHeader.Bits)))
	fmt.Println("IntToHex(int64(nonce))", IntToHex(int64(nonce)))
	data := bytes.Join(
		[][]byte{
			pow.block.BlockHeader.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.BlockHeader.Timestamp),
			IntToHex(int64(pow.block.BlockHeader.Bits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block\n")
CalculateNonce:
	for nonce < maxNonce {
		select {
		case <-receiveBlockChan:
			fmt.Printf("\rThe block has been mined!\n")
			return -1, nil
		default:
			data := pow.prepareData(nonce)

			hash = sha256.Sum256(data)
			fmt.Printf("\r%x", hash)
			hashInt.SetBytes(hash[:])

			if hashInt.Cmp(pow.target) == -1 {
				break CalculateNonce
			} else {
				nonce++
			}
		}
	}
	fmt.Printf("\nNew block is mined!\n")

	return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	fmt.Println(pow.block.BlockHeader)

	data := pow.prepareData(pow.block.BlockHeader.Nonce)
	fmt.Println("pow.prepareData(pow.block.BlockHeader.Nonce)", data)
	hash := sha256.Sum256(data)
	fmt.Println("sha256.Sum256(data)", hash)
	hashInt.SetBytes(hash[:])
	fmt.Println("hashInt.SetBytes(hash[:])", hashInt)

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
