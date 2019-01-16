package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var blockChanFromBlockChainManagerToServer chan *Block = make(chan *Block, 20)
var blockChanFromBlockChainManagerToMempool chan *Block = make(chan *Block, 20)

var blocksHashChanFromBlockChainManagerToServer chan [][]byte = make(chan [][]byte, 20)

var blockHashChanFromBlockChainManagerToServer chan *Block = make(chan *Block, 20)

// BlockChainManager ...
type BlockChainManager struct {
	BlockHeader    *BlockHeader
	Hash           []byte
	TransactionNum int
	Transactions   []*Transaction
	Height         int64
	mtx            *sync.Mutex
}

// NewBlockChainManager returns a new BlockChainManager
func NewBlockChainManager(block *Block) *BlockChainManager {
	return &BlockChainManager{
		block.BlockHeader,
		block.Hash,
		block.TransactionNum,
		block.Transactions,
		block.Height,
		new(sync.Mutex),
	}
}

// BlockChainProcessor process blockchain, receive block from server, receive txs from mempool,
// store blockchains in db, returns blocks to server
func (bcm *BlockChainManager) BlockChainProcessor() {
	select {
	case block := <-blockChanFromServerToBlockChainManager:
		go bcm.ValidateBlockIsValidAndAdd(block)
	case txs := <-txChanFromMempoolToBlockChainManager:
		go bcm.MineBlock(txs)
	}
}

// ValidateBlockIsValidAndAdd refers BlockChainManager will validate the received block
func (bcm *BlockChainManager) ValidateBlockIsValidAndAdd(block *Block) {

	// if PoW is not valid, do nothing
	if !block.VerifyPoW() {
		fmt.Printf("The block is invalid!\n")
		return
	}

	height := bcm.GetHeight()

	// if block is smaller than current height and last 6 blocks, we think the block is usefulless
	if block.Height+termValidityOfBlock <= height {
		return
	}

	// if block's height is bigger than current height, add the block into blockchain
	if block.Height > height {
		//receiveBlockChan <- true
	}

	bcm.AddBlock(block)

	UTXOSet := UTXOSet{bcm.Hash}
	UTXOSet.Reindex()
}

// MineBlock ...
func (bcm *BlockChainManager) MineBlock(txs []*Transaction) {
	cbTx := NewCoinbaseTX(miningAddress, "")
	txs = append(txs, cbTx)

	for _, tx := range txs {
		// TODO: ignore transaction if it's not valid
		if bcm.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	lastBlock := bcm.GetLastBlock()
	lastHash, lastHeight := lastBlock.Hash, lastBlock.Height

	newBlock := NewBlock(txs, lastHash, lastHeight+1)

	bcm.AddBlock(newBlock)

	if newBlock == nil {
		return
	}

	UTXOSet := UTXOSet{bcm.Hash}
	UTXOSet.Reindex()

	mp.DeleteTxs(txs)
}

// GetHeight returns height stored in BlockChainManager
func (bcm *BlockChainManager) GetHeight() int64 {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	height := bcm.Height
	return height
}

// GetVersion returns version stored in BlockChainManager
func (bcm *BlockChainManager) GetVersion() int {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	version := bcm.BlockHeader.Version
	return version

}

// GetHash returns the hash of the latest block
func (bcm *BlockChainManager) GetHash() []byte {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	hash := bcm.Hash[:]
	return hash[:]
}

// GetBlockByHash finds a block by its hash and returns it
func (bcm *BlockChainManager) GetBlockByHash(blockHash []byte) (*Block, error) {
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
func (bcm *BlockChainManager) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bcm.Iterator(bcm.Hash)

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

// GetLastBlock returns the latest block
func (bcm *BlockChainManager) GetLastBlock() *Block {
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

// Iterator returns a BlockchainIterat
func (bcm *BlockChainManager) Iterator(lastHash []byte) *BlockchainIterator {
	bci := &BlockchainIterator{lastHash}

	return bci
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bcm *BlockChainManager) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bcm.Iterator(bcm.Hash)

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

// VerifyTransaction verifies transaction input signatures
func (bcm *BlockChainManager) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bcm.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// SignTransaction signs inputs of a Transaction
func (bcm *BlockChainManager) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bcm.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// FindTransaction finds a transaction by its ID
func (bcm *BlockChainManager) FindTransaction(ID []byte) (Transaction, error) {
	bci := bcm.Iterator(bcm.Hash)

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

// AddBlock saves the block into the blockchain
func (bcm *BlockChainManager) AddBlock(newBlock *Block) {
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

		newBlockData := newBlock.Serialize()

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
			bci := bcm.Iterator(bcm.Hash)
			for {
				block := bci.Next()
				if CompareHash(block.Hash, newBlock.BlockHeader.PrevBlockHash) && block.Height+1 == newBlock.Height {
					err := b.Put(newBlock.Hash, newBlockData)
					if err != nil {
						log.Panic(err)
					}

					newSideChainIndex := bcm.NewSideChainIndex(b)

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
			if CompareHash(newBlock.BlockHeader.PrevBlockHash, lastBlock.Hash) {
				err := b.Put(newBlock.Hash, newBlockData)
				if err != nil {
					log.Panic(err)
				}
				err = b.Put([]byte("l"), newBlock.Hash)
				if err != nil {
					log.Panic(err)
				}
				bcm.Hash = newBlock.Hash
			}
		}

		// after newBlock is added ,check the orphan pool whether the orpahan block can be add into blockchain
	UpdateOrphanBlock:
		c := o.Cursor()
		for hashByte, blockDataByte := c.First(); hashByte != nil; hashByte, blockDataByte = c.Next() {
			block := DeserializeBlock(blockDataByte)
			if CompareHash(block.BlockHeader.PrevBlockHash, newBlock.Hash) {
				err := b.Put(hashByte, blockDataByte)
				if err != nil {
					log.Panic(err)
				}

				o.Delete(hashByte)

				if CompareHash(b.Get([]byte("l")), newBlock.Hash) {
					err = b.Put([]byte("l"), block.Hash)
					if err != nil {
						log.Panic(err)
					}
				} else {
					for i := 1; i < isAllowedSideChainNum; i++ {
						if CompareHash(b.Get([]byte(IntToByte(i))), newBlock.Hash) {
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

// NewSideChainIndex returns the index of the new side chain
func (bcm *BlockChainManager) NewSideChainIndex(b *bolt.Bucket) int {
	for i := 1; i <= isAllowedSideChainNum; i++ {
		if b.Get([]byte(IntToByte(i))) == nil {
			return i
		}
	}

	newIndex := bcm.GetCurrentOldestSideChain(b)
	return newIndex
}

// GetCurrentOldestSideChain get current oldest side chain
func (bcm *BlockChainManager) GetCurrentOldestSideChain(b *bolt.Bucket) int {
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

			for k := range tmpIndex {
				delete(tmpIndex, k)
			}

			tmpIndex[i] = i
		}
	}

	sideChainIndex = tmpIndex

	if len(sideChainIndex) == 1 {
		for k := range sideChainIndex {
			bcm.DeleteOldestSideChain(b, k)
			return k
		}
	}

	// 再比每条侧链分叉块的块高度

	lowestforkBlockHeight := lastBlock.Height

	tmpIndex = sideChainIndex

	for k := range sideChainIndex {
		forkBlockHeight := bcm.GetForkBlockHeight(b, k)

		if forkBlockHeight > lowestforkBlockHeight {
			continue
		}

		if forkBlockHeight == lowestforkBlockHeight {
			tmpIndex[k] = k
		}

		if forkBlockHeight < lowestforkBlockHeight {
			lowestforkBlockHeight = forkBlockHeight

			for k := range tmpIndex {
				delete(tmpIndex, k)
			}

			tmpIndex[k] = k
		}
	}

	sideChainIndex = tmpIndex

	if len(sideChainIndex) == 1 {
		for k := range sideChainIndex {
			bcm.DeleteOldestSideChain(b, k)
			return k
		}
	}

	// 最后比每条侧链最后块的时间戳

	oldestTimestamp := time.Now().Unix()

	newIndex := 0

	for k := range sideChainIndex {
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

	bcm.DeleteOldestSideChain(b, newIndex)
	return newIndex
}

// DeleteOldestSideChain delete side chain which is the oldest
func (bcm *BlockChainManager) DeleteOldestSideChain(b *bolt.Bucket, sideChainIndex int) {
	forkHeight := bcm.GetForkBlockHeight(b, sideChainIndex)

	bcsi := bcm.Iterator(b.Get([]byte(IntToByte(sideChainIndex))))
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

// GetForkBlockHeight get fork height of blockchain
func (bcm *BlockChainManager) GetForkBlockHeight(b *bolt.Bucket, sideChainIndex int) int64 {
	var height int64 = 0

	mainChainMap := []string{}
	sideChainMap := []string{}

	bci := bcm.Iterator(bcm.Hash)
	for {
		block := bci.Next()
		mainChainMap = append(mainChainMap, string(block.Hash))

		if len(block.BlockHeader.PrevBlockHash) == 0 {
			break
		}
	}

	bcsi := bcm.Iterator(b.Get([]byte(IntToByte(sideChainIndex))))
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
