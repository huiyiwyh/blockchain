package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

// BlockchainManager ...
type BlockchainManager struct {
	BlockHeader       *BlockHeader
	Hash              []byte
	TransactionNum    int
	Transactions      []*Transaction
	Height            int64
	SidechanTimestamp map[int64]string
	IsMining          bool
	mtx               *sync.Mutex
}

// BlockchainManagerInfo ...
type BlockchainManagerInfo struct {
	BlockHeader    *BlockHeader
	Hash           []byte
	TransactionNum int
	Transactions   []*Transaction
	Height         int64
}

// NewBlockchainManager returns a new BlockchainManager
func NewBlockchainManager() *BlockchainManager {
	block, _ := LoadTopBlock()

	return &BlockchainManager{
		block.BlockHeader,
		block.Hash,
		block.TransactionNum,
		block.Transactions,
		block.Height,
		make(map[int64]string),
		false,
		new(sync.Mutex),
	}
}

// BlockchainProcessor process Blockchain, receive block from server, receive txs from mempool,
// store Blockchains in db, returns blocks to server
func (bcm *BlockchainManager) Processor() {
	for {
		select {
		case block := <-SToBBlock:
			go bcm.validateBlockIsValidAndAdd(block)
		case txs := <-MToBTxs:
			go bcm.mineBlock(txs)
		case <-CToBGetBCMI:
			go bcm.returnCliBlockchainManagerInfo()
		case <-SToBGetBCMI:
			go bcm.returnServerBlockchainManagerInfo()
		case <-MToBGetBCMI:
			go bcm.returnMempoolBlockchainManagerInfo()
		case blockByHash := <-SToBGetBlockByHash:
			go bcm.returnServerBlockbyHash(blockByHash)
		case blocksHash := <-SToBGetBlocksHash:
			go bcm.returnServerBlocksHash(blocksHash)
		}
	}
}

// ReturnServerBlockchainManagerInfo returns BlockchainManagerinfo to Server
func (bcm *BlockchainManager) returnServerBlockchainManagerInfo() {
	nbcmi := newBlockchainManagerInfo(bcm)
	BToSBCMI <- nbcmi
}

// ReturnCliBlockchainManagerInfo returns BlockchainManagerinfo to cli
func (bcm *BlockchainManager) returnCliBlockchainManagerInfo() {
	nbcmi := newBlockchainManagerInfo(bcm)
	BToCBCMI <- nbcmi
}

// ReturnMempoolBlockchainManagerInfo returns BlockchainManagerinfo to mempool
func (bcm *BlockchainManager) returnMempoolBlockchainManagerInfo() {
	nbcmi := newBlockchainManagerInfo(bcm)
	BToMBCMI <- nbcmi
}

// returnServerBlockbyHash returns BlockByHash to server
func (bcm *BlockchainManager) returnServerBlockbyHash(blockByHash *BlockByHash) {
	block, err := bcm.getBlockByHash(blockByHash.Hash)
	if err != nil {
		fmt.Println(err)
	}
	BToSBlockByHash <- &BlockByHash{blockByHash.NodeFrom, blockByHash.Hash, block}
}

// returnServerBlocksHash returns BlocksHash to server
func (bcm *BlockchainManager) returnServerBlocksHash(blocksHash *BlocksHash) {
	BToSBlocksHash <- &BlocksHash{blocksHash.NodeFrom, bcm.getBlocksHash()}
}

// ValidateBlockIsValidAndAdd refers BlockchainManager will validate the received block
func (bcm *BlockchainManager) validateBlockIsValidAndAdd(block *Block) {

	// if PoW is not valid, do nothing
	if !block.VerifyPoW() {
		fmt.Printf("The block is invalid!\n")
		return
	}

	height := bcm.getHeight()

	// if block is smaller than current height and last 6 blocks, we think the block is usefulless
	if block.Height+termValidityOfBlock <= height {
		return
	}

	// if block's height is bigger than current height, add the block into Blockchain
	if block.Height > height && bcm.getIsMining() {
		ReceivedBlock <- &Notification{}
	}

	bcm.addBlock(block)

	UTXOSet := UTXOSet{bcm.Hash}
	UTXOSet.Reindex()
}

// MineBlock ...
func (bcm *BlockchainManager) mineBlock(txs []*Transaction) {
	cbTx := NewCoinbaseTX(miningAddress, "")
	txs = append(txs, cbTx)

	lastBlock := bcm.getLastBlock()
	lastHash, lastHeight := lastBlock.Hash, lastBlock.Height

	bcm.changeIsMining(true)
	newBlock := NewBlock(txs, lastHash, lastHeight+1)
	bcm.changeIsMining(false)
	if newBlock == nil {
		return
	}

	bcm.addBlock(newBlock)

	UTXOSet := UTXOSet{bcm.Hash}
	UTXOSet.Reindex()

	BToMTxs <- txs
	BToSBlock <- newBlock
}

// GetHeight returns height stored in BlockchainManager
func (bcm *BlockchainManager) getHeight() int64 {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	height := bcm.Height
	return height
}

// GetVersion returns version stored in BlockchainManager
func (bcm *BlockchainManager) getVersion() int {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	version := bcm.BlockHeader.Version
	return version

}

// GetHash returns the hash of the latest block
func (bcm *BlockchainManager) getHash() []byte {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	hash := bcm.Hash[:]
	return hash[:]
}

//
func (bcm *BlockchainManager) changeIsMining(isMining bool) {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	bcm.IsMining = isMining
}

// getIsMining returns IsMining of the bcm
func (bcm *BlockchainManager) getIsMining() bool {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	isMining := bcm.IsMining
	return isMining
}

// GetBlockByHash finds a block by its hash and returns it
func (bcm *BlockchainManager) getBlockByHash(blockHash []byte) (*Block, error) {
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

// GetBlocksHash returns a list of hashes of all the blocks in the chain
func (bcm *BlockchainManager) getBlocksHash() [][]byte {
	var blocks [][]byte

	bc := &Blockchain{bcm.Hash}
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

// GetLastBlock returns the latest block
func (bcm *BlockchainManager) getLastBlock() *Block {
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

// AddBlock saves the block into the Blockchain
func (bcm *BlockchainManager) addBlock(newBlock *Block) {
	db, err := bolt.Open(blockchaindbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	newBlockData := newBlock.Serialize()

	err = db.Update(func(tx *bolt.Tx) error {
		falg := false
		b := tx.Bucket([]byte(blocksBucket))
		o := tx.Bucket([]byte(orphanBlocksBucket))

		blockInBlocksHash := b.Get(newBlock.Hash)
		blockInOrphanBlocksHash := o.Get(newBlock.Hash)

		if blockInBlocksHash != nil || blockInOrphanBlocksHash != nil {
			// fmt.Printf("the block %x is in the main chain or orphan pool\n", newBlock.Hash)
			return nil
		}

		lastHash := b.Get([]byte("l"))

		// as orphanblock to be added
		c := b.Cursor()
		for hashByte, _ := c.First(); hashByte != nil; hashByte, _ = c.Next() {
			if bytes.Compare(newBlock.BlockHeader.PrevBlockHash, hashByte) == 0 {
				err := b.Put(newBlock.Hash, newBlockData)
				if err != nil {
					log.Panic(err)
				}

				if bytes.Compare(hashByte, lastHash) == 0 {
					err = b.Put([]byte("l"), newBlock.Hash)
					if err != nil {
						log.Panic(err)
					}
					// fmt.Printf("block %x as mainchain to be added\n", newBlock.Hash)
					falg = true
					break
				}

				err = b.Put(newBlock.Hash, newBlockData)
				if err != nil {
					log.Panic(err)
				}

				newTimestamp := bcm.getNewSidechainTimestamp(hashByte)

				err = b.Put([]byte(newTimestamp), newBlock.Hash)
				if err != nil {
					log.Panic(err)
				}
				// fmt.Printf("block %x as sidechain to be added\n", newBlock.Hash)
				falg = true
				break
			}
		}

		if !falg {
			err := o.Put(newBlock.Hash, newBlockData)
			if err != nil {
				log.Panic(err)
			}
			// fmt.Printf("block %x as orphanblock to be added\n", newBlock.Hash)

			return nil
		}

		// fmt.Println("UpdateOrphanBlock")
		// after newBlock is added ,check the orphan pool whether the orpahan block can be add into Blockchain
	UpdateOrphanBlock:
		oc := o.Cursor()
		for hashByte, blockDataByte := oc.First(); hashByte != nil; hashByte, blockDataByte = oc.Next() {
			block := DeserializeBlock(blockDataByte)
			if bytes.Compare(block.BlockHeader.PrevBlockHash, newBlock.Hash) == 0 {
				err := b.Put(hashByte, blockDataByte)
				if err != nil {
					log.Panic(err)
				}
				o.Delete(hashByte)

				if bytes.Compare(b.Get([]byte("l")), newBlock.Hash) == 0 {
					err = b.Put([]byte("l"), block.Hash)
					if err != nil {
						log.Panic(err)
					}
					// fmt.Printf("block %x as mainchain to be added\n", block.Hash)
				} else {
					timestamp := bcm.getNewSidechainTimestamp(newBlock.Hash)
					err = b.Put([]byte(timestamp), block.Hash)
					if err != nil {
						log.Panic(err)
					}
					// fmt.Printf("block %x as sidechain to be added\n", block.Hash)

					// compare height of the sidechain and mainchain
					// if sidechain's height higher than the height of mainchain, change the tag of the hash
					lastHash = b.Get([]byte("l"))
					lastBlockData := b.Get(lastHash)
					lastBlock := DeserializeBlock(lastBlockData)

					sideLastHash := b.Get([]byte(timestamp))
					sideBlockData := b.Get([]byte(sideLastHash))
					sideLastBlock := DeserializeBlock(sideBlockData)

					if lastBlock.Height < sideLastBlock.Height {
						err = b.Put([]byte("l"), sideLastBlock.Hash)
						if err != nil {
							log.Panic(err)
						}
						err = b.Put([]byte(timestamp), lastBlock.Hash)
						if err != nil {
							log.Panic(err)
						}
						// fmt.Println("sidechain change to mainchain")
					}
					break
				}

				newBlock = block
				goto UpdateOrphanBlock
			}
		}

		hash := b.Get([]byte("l"))
		blockDataByte := b.Get(hash)
		block := DeserializeBlock(blockDataByte)
		bcm.updateBlockchainManager(block)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	// fmt.Println("end")
}

// GetCurrentOldestSideChain get current oldest side chain
func (bcm *BlockchainManager) getCurrentOldestSideChain(b *bolt.Bucket) int {
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
			bcm.deleteOldestSideChain(b, k)
			return k
		}
	}

	// 再比每条侧链分叉块的块高度

	lowestforkBlockHeight := lastBlock.Height

	tmpIndex = sideChainIndex

	for k := range sideChainIndex {
		forkBlockHeight := bcm.getForkBlockHeight(b, k)

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
			bcm.deleteOldestSideChain(b, k)
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

	bcm.deleteOldestSideChain(b, newIndex)
	return newIndex
}

// DeleteOldestSideChain delete side chain which is the oldest
func (bcm *BlockchainManager) deleteOldestSideChain(b *bolt.Bucket, sideChainIndex int) {
	forkHeight := bcm.getForkBlockHeight(b, sideChainIndex)
	bc := &Blockchain{bcm.Hash}

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

// GetForkBlockHeight get fork height of Blockchain
func (bcm *BlockchainManager) getForkBlockHeight(b *bolt.Bucket, sideChainIndex int) int64 {
	var height int64 = 0

	mainChainMap := []string{}
	sideChainMap := []string{}

	bc := &Blockchain{bcm.Hash}
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

// GetBlockchainManagerInfo ...
func newBlockchainManagerInfo(bcm *BlockchainManager) *BlockchainManagerInfo {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	nbcm := &BlockchainManagerInfo{}
	nbcm.Hash = bcm.Hash
	nbcm.Height = bcm.Height
	nbcm.TransactionNum = bcm.TransactionNum
	nbcm.Transactions = bcm.Transactions
	nbcm.BlockHeader = bcm.BlockHeader

	return nbcm
}

func (bcm *BlockchainManager) updateBlockchainManager(block *Block) {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	bcm.Hash = block.Hash
	bcm.Height = block.Height
	bcm.TransactionNum = block.TransactionNum
	bcm.Transactions = block.Transactions
	bcm.BlockHeader = block.BlockHeader
}

func (bcm *BlockchainManager) getNewSidechainTimestamp(blockHash []byte) string {
	timestamp := bcm.checkSidechainTimestamp(string(blockHash))

	if timestamp == -1 {
		timestamp = time.Now().Unix()
	}

	return strconv.FormatInt(timestamp, 10)
}

func (bcm *BlockchainManager) checkSidechainTimestamp(blockID string) int64 {
	timestamps := bcm.getSidechainTimestamp()

	for timestamp, blockId := range timestamps {
		if blockID == blockId {
			return timestamp
		}
	}

	return -1
}

func (bcm *BlockchainManager) deleteSidechainNotif(timestamp int64) {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	delete(bcm.SidechanTimestamp, timestamp)
}

func (bcm *BlockchainManager) addSidechainNotif(timestamp int64, blockID string) {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	bcm.SidechanTimestamp[timestamp] = blockID
}

func (bcm *BlockchainManager) getSidechainTimestamp() map[int64]string {
	bcm.mtx.Lock()
	defer bcm.mtx.Unlock()

	return bcm.SidechanTimestamp
}
