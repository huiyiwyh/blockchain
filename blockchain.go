package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

const blockchaindbFile = "blockchain.db"
const blocksBucket = "blocks"
const orphanBlocksBucket = "orphanblocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const isAllowedSideChainNum = 5

var nextSideChainIndex int = 1

// Blockchain implements interactions with a DB
type Blockchain struct {
	tip []byte
	// db  *bolt.DB
}

func CreateOrLoadBlockChaindb() {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
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

		err = b.Put(genesis.Hash, Serialize(genesis))
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

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(newBlock *Block) {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		o := tx.Bucket([]byte(orphanBlocksBucket))

		blockInBlocksHash := b.Get(newBlock.Hash)
		blockInOrphanBlocksHash := o.Get(newBlock.Hash)

		if blockInBlocksHash != nil || blockInOrphanBlocksHash != nil {
			return nil
		}

		newBlockData := Serialize(newBlock)

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		// as orphanblock to be added
		if newBlock.Height > lastBlock.Height+1 {
			err := o.Put(newBlock.Hash, newBlockData)
			if err != nil {
				log.Panic(err)
			}

			return nil
		}

		// as sidechain to be added
		if newBlock.Height < lastBlock.Height {
			bci := bc.Iterator(bc.tip)
			for {
				block := bci.Next()
				if compareHash(block.Hash, newBlock.BlockHeader.PrevBlockHash) && block.Height+1 == newBlock.Height {
					err := b.Put(newBlock.Hash, newBlockData)
					if err != nil {
						log.Panic(err)
					}

					newSideChainIndex := bc.NewSideChainIndex(b)

					err = b.Put([]byte(IntToByte(newSideChainIndex)), newBlock.Hash)
					if err != nil {
						log.Panic(err)
					}

					break
				}
			}
		}

		// as mainchain to be added
		if newBlock.Height == lastBlock.Height+1 {
			if compareHash(newBlock.BlockHeader.PrevBlockHash, lastBlock.Hash) {
				err := b.Put(newBlock.Hash, newBlockData)
				if err != nil {
					log.Panic(err)
				}
				err = b.Put([]byte("l"), newBlock.Hash)
				if err != nil {
					log.Panic(err)
				}
				bc.tip = newBlock.Hash
			}
		}

		// after newBlock is added ,check the orphan pool whether the orpahan block can be add into blockchain
	UpdateOrphanBlock:
		c := o.Cursor()
		for hashByte, blockDataByte := c.First(); hashByte != nil; hashByte, blockDataByte = c.Next() {
			block := DeserializeBlock(blockDataByte)
			if compareHash(block.BlockHeader.PrevBlockHash, newBlock.Hash) {
				err := b.Put(hashByte, blockDataByte)
				if err != nil {
					log.Panic(err)
				}

				o.Delete(hashByte)

				if compareHash(b.Get([]byte("l")), newBlock.Hash) {
					err = b.Put([]byte("l"), block.Hash)
					if err != nil {
						log.Panic(err)
					}
				} else {
					for i := 1; i < isAllowedSideChainNum; i++ {
						if compareHash(b.Get([]byte(IntToByte(i))), newBlock.Hash) {
							err = b.Put([]byte(IntToByte(i)), block.Hash)
							if err != nil {
								log.Panic(err)
							}

							// compare height of the sidechain and mainchain
							// if sidechain's height higher than the height of mainchain, change the tag of the hash
							lastHash = b.Get([]byte("l"))
							lastBlockData := b.Get(lastHash)
							lastBlock := DeserializeBlock(lastBlockData)

							sideLastHash := b.Get([]byte(IntToByte(i)))
							sideBlockData := b.Get([]byte(sideLastHash))
							sideLastBlock := DeserializeBlock(sideBlockData)

							if lastBlock.Height < sideLastBlock.Height {
								err = b.Put([]byte("l"), sideLastBlock.Hash)
								if err != nil {
									log.Panic(err)
								}
								err = b.Put([]byte(IntToByte(i)), lastBlock.Hash)
								if err != nil {
									log.Panic(err)
								}
							}
							break
						}
					}
				}
				newBlock = block
				goto UpdateOrphanBlock
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator(bc.tip)

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if block.BlockHeader.PrevBlockHash == nil {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator(bc.tip)

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator(lastHash []byte) *BlockchainIterator {
	bci := &BlockchainIterator{lastHash}

	return bci
}

// GetLastBlock returns the latest block
func (bc *Blockchain) GetLastBlock() *Block {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var lastBlock *Block

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock
}

// GetVersion returns the version of the latest block
func (bc *Blockchain) GetVersion() int {
	return bc.GetLastBlock().BlockHeader.Version
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	return bc.GetLastBlock().Height
}

// GetBlockHash returns the hash of the latest block
func (bc *Blockchain) GetBlockHash() []byte {
	return bc.GetLastBlock().Hash
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlockByHash(blockHash []byte) (*Block, error) {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	var block *Block

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found.")
		}

		block = DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator(bc.tip)

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	for _, tx := range transactions {
		// TODO: ignore transaction if it's not valid
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	lastBlock := bc.GetLastBlock()
	lastHash, lastHeight := lastBlock.Hash, lastBlock.Height

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)
	if newBlock == nil {
		return nil
	}

	bc.AddBlock(newBlock)

	fmt.Println("after add block to blockchain")

	return newBlock
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) NewSideChainIndex(b *bolt.Bucket) int {
	for i := 1; i <= isAllowedSideChainNum; i++ {
		if b.Get([]byte(IntToByte(i))) == nil {
			return i
		}
	}

	newIndex := bc.GetCurrentOldestSideChain(b)
	return newIndex
}

func (bc *Blockchain) GetCurrentOldestSideChain(b *bolt.Bucket) int {
	lastHash := b.Get([]byte("l"))
	blockData := b.Get(lastHash)
	lastBlock := DeserializeBlock(blockData)

	// 先比每条侧链最后块的块高度

	lowestSideChainHeight := lastBlock.Height
	sideChainIndex := map[int]int{}
	tmpIndex := sideChainIndex

	for i := 1; i <= isAllowedSideChainNum; i++ {
		sideLastHash := b.Get([]byte(IntToByte(i)))
		sideBlockData := b.Get([]byte(sideLastHash))
		sideLastBlock := DeserializeBlock(sideBlockData)

		if sideLastBlock.Height > lowestSideChainHeight {
			continue
		}

		if sideLastBlock.Height == lowestSideChainHeight {
			tmpIndex[i] = i
		}

		if sideLastBlock.Height < lowestSideChainHeight {
			lowestSideChainHeight = sideLastBlock.Height

			for k, _ := range tmpIndex {
				delete(tmpIndex, k)
			}

			tmpIndex[i] = i
		}
	}

	sideChainIndex = tmpIndex

	if len(sideChainIndex) == 1 {
		for k, _ := range sideChainIndex {
			bc.DeleteOldestSideChain(b, k)
			return k
		}
	}

	// 再比每条侧链分叉块的块高度

	lowestforkBlockHeight := lastBlock.Height

	tmpIndex = sideChainIndex

	for k, _ := range sideChainIndex {
		forkBlockHeight := bc.GetForkBlockHeight(b, k)

		if forkBlockHeight > lowestforkBlockHeight {
			continue
		}

		if forkBlockHeight == lowestforkBlockHeight {
			tmpIndex[k] = k
		}

		if forkBlockHeight < lowestforkBlockHeight {
			lowestforkBlockHeight = forkBlockHeight

			for k, _ := range tmpIndex {
				delete(tmpIndex, k)
			}

			tmpIndex[k] = k
		}
	}

	sideChainIndex = tmpIndex

	if len(sideChainIndex) == 1 {
		for k, _ := range sideChainIndex {
			bc.DeleteOldestSideChain(b, k)
			return k
		}
	}

	// 最后比每条侧链最后块的时间戳

	oldestTimestamp := time.Now().Unix()

	newIndex := 0

	for k, _ := range sideChainIndex {
		sideLastHash := b.Get([]byte(IntToByte(k)))
		sideBlockData := b.Get([]byte(sideLastHash))
		sideLastBlock := DeserializeBlock(sideBlockData)

		if sideLastBlock.BlockHeader.Timestamp > oldestTimestamp {
			continue
		}

		if sideLastBlock.BlockHeader.Timestamp < oldestTimestamp {
			oldestTimestamp = sideLastBlock.BlockHeader.Timestamp
			newIndex = k
		}
	}

	bc.DeleteOldestSideChain(b, newIndex)
	return newIndex
}

func (bc *Blockchain) DeleteOldestSideChain(b *bolt.Bucket, sideChainIndex int) {
	forkHeight := bc.GetForkBlockHeight(b, sideChainIndex)

	bcsi := bc.Iterator(b.Get([]byte(IntToByte(sideChainIndex))))
	for {
		block := bcsi.Next()
		if block.Height > forkHeight {
			b.Delete([]byte(block.Hash))
		}

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}
	b.Delete([]byte(IntToByte(sideChainIndex)))
}

func (bc *Blockchain) GetForkBlockHeight(b *bolt.Bucket, sideChainIndex int) int {
	height := 0

	mainChainMap := []string{}
	sideChainMap := []string{}

	bci := bc.Iterator(bc.tip)
	for {
		block := bci.Next()
		mainChainMap = append(mainChainMap, string(block.Hash))

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	bcsi := bc.Iterator(b.Get([]byte(IntToByte(sideChainIndex))))
	for {
		block := bcsi.Next()
		sideChainMap = append(sideChainMap, string(block.Hash))

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	for i, j := len(mainChainMap)-1, len(sideChainMap)-1; i > -1 && j > -1; i, j = i+1, j-1 {
		if mainChainMap[i] != sideChainMap[j] {
			break
		}
		height++
	}

	if height == 0 {
		log.Panic("The genesis block is different!")
	}

	return height - 1
}
