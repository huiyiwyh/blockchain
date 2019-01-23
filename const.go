package main

// protocol refers the connection protocol
const protocol = "tcp"

// commandLength refers the command's length
const commandLength = 12

// termValidityOfBlock refers the block
const termValidityOfBlock int64 = 6

// blockversion refers the blockversion
const blockversion = byte(0x00)

// addressChecksumLen defines the addressChecksumLen
const addressChecksumLen = 4

// MempoolMaxTxs defers maxTxs num in mempool
const mempoolMaxTxs = 2

// BlockchaindbFile defers dbname
const blockchaindbFile = "blockchain.db"

// blocksBucket defines the bucket of block in Blockchain
const blocksBucket = "blocks"

// orphanBlocksBucket defines the bucket of orphan block
const orphanBlocksBucket = "orphanblocks"

// genesisCoinbaseData defines the genesisCoin baseData
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// isAllowedSideChainNum defines how many side chan will be saved
const isAllowedSideChainNum = 5

// subsidy defines when a block mind ,the miner can receive 20
const subsidy = 20

// port defines which port to transport informations
const port = "3000"
