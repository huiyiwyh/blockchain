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
const MempoolMaxTxs = 2
