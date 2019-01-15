package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
)

const protocol = "tcp"
const commandLength = 12
const termValidityOfBlock = 6

var localNodeAddress string
var miningAddress string

var blocksInTransit = [][]byte{}

// receiveBlockChan ...
var receiveBlockChan chan bool = make(chan bool)

// Mining ...
var Mining bool = false

var mp = &Mempool{make(map[string]*Transaction), 0, new(sync.Mutex)}

// Mempool ...
type Mempool struct {
	mempool map[string]*Transaction
	txNums  int
	Lock    *sync.Mutex
}

type block struct {
	NodeFrom string
	Block    []byte
}

type getblocks struct {
	NodeFrom string
}

type getdata struct {
	NodeFrom string
	Type     string
	ID       []byte
}

type inv struct {
	NodeFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	NodeFrom    string
	Transaction []byte
}

type verzion struct {
	NodeFrom   string
	Version    int
	BestHeight int
}

type wallet struct {
	NodeFrom string
	Wallets  []string
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

func extractCommand(request []byte) []byte {
	return request[:commandLength]
}

func requestBlocks() {

	//sendGetBlocks(node)
}

func sendBlock(nodeTo string, b *Block) {
	data := block{localNodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(nodeTo, request)
}

func sendData(nodeTo string, data []byte) {
	conn, err := net.Dial(protocol, nodeTo)
	if err != nil {
		fmt.Printf("%s is not available\n", nodeTo)

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(nodeTo, kind string, items [][]byte) {
	inventory := inv{localNodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(nodeTo, request)
}

func sendGetBlocks(nodeTo string) {
	payload := gobEncode(getblocks{localNodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(nodeTo, request)
}

func sendGetData(nodeTo, kind string, id []byte) {
	payload := gobEncode(getdata{localNodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(nodeTo, request)
}

func sendTx(nodeTo string, tnx *Transaction) {
	data := tx{localNodeAddress, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(nodeTo, request)
}

func sendVersion(nodeTo string, blockVersion int, bestHeight int) {
	payload := gobEncode(verzion{localNodeAddress, blockVersion, bestHeight})

	request := append(commandToBytes("version"), payload...)

	sendData(nodeTo, request)
}

func handleBlock(request []byte) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	bc := NewBlockchain()

	fmt.Printf("Recevied a new block!\n")

	pow := NewProofOfWork(block)
	if !pow.Validate() {
		fmt.Printf("The block is invalid!\n")
		return
	}

	if block.Height+termValidityOfBlock <= bc.GetBestHeight() {
		return
	}

	if block.Height > bc.GetBestHeight() {
		receiveBlockChan <- true
	}

	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	txs := block.Transactions

	MempoolDeleteTxs(txs)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.NodeFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s from %s\n", len(payload.Items), payload.Type, payload.NodeFrom)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.NodeFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		mp.Lock.Lock()
		if mp.mempool[hex.EncodeToString(txID)] == nil {
			sendGetData(payload.NodeFrom, "tx", txID)
		}
		mp.Lock.Unlock()
	}
}

func handleGetBlocks(request []byte) {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received getblocks command from %s\n", payload.NodeFrom)

	bc := NewBlockchain()

	blocks := bc.GetBlockHashes()
	sendInv(payload.NodeFrom, "block", blocks)
}

func handleGetData(request []byte) {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received getdata command from %s\n", payload.NodeFrom)

	bc := NewBlockchain()

	if payload.Type == "block" {
		block, err := bc.GetBlockByHash([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.NodeFrom, block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)

		mp.Lock.Lock()
		tx := mp.mempool[txID]
		mp.Lock.Unlock()

		sendTx(payload.NodeFrom, tx)
	}
}

func handleTx(request []byte) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("\rReceived tx command from %s\n", payload.NodeFrom)
	txData := payload.Transaction
	tx := DeserializeTransaction(txData)

	MempoolAddTxs(tx)

	// for _, node := range nodesInfo[localNodeAddress].Nodes {
	// 	if node != localNodeAddress && node != payload.NodeFrom {
	// 		sendInv(node, "tx", [][]byte{tx.ID})
	// 	}
	// }
	fmt.Println("txs in mempool's number :", MempoolGetTxNums())

	if MempoolGetTxNums() > 1 {
	MineTransactions:
		bc := NewBlockchain()
		txs := MempoolVerifyTxs(bc)

		if len(txs) == 0 {
			fmt.Printf("All transactions are added into block! Waiting for new ones...\n")
			return
		}

		// create coinbaseTx and add into mempool
		cbTx := NewCoinbaseTX(miningAddress, "")
		txs = append(txs, cbTx)

		Mining = true
		newBlock := bc.MineBlock(txs)
		Mining = false

		if newBlock == nil {
			return
		}

		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()

		MempoolDeleteTxs(txs)

		// for _, node := range nodesInfo[localNodeAddress].Nodes {
		// 	if node != localNodeAddress {
		// 		sendInv(node, "block", [][]byte{newBlock.Hash})
		// 	}
		// }

		if MempoolGetTxNums() > 0 {
			goto MineTransactions
		}
	}
}

func handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received version command from %s\n", payload.NodeFrom)

	bc := NewBlockchain()

	if payload.Version < bc.GetVersion() {
		return
	}

	myBestHeight := bc.GetBestHeight()
	myBlockVersion := bc.GetVersion()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight > foreignerBestHeight {
		sendVersion(payload.NodeFrom, myBlockVersion, myBestHeight)
	}

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.NodeFrom)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])

	switch command {
	case "block":
		handleBlock(request)
	case "inv":
		handleInv(request)
	case "getblocks":
		handleGetBlocks(request)
	case "getdata":
		handleGetData(request)
	case "tx":
		handleTx(request)
	case "version":
		handleVersion(request)
	default:
		fmt.Printf("Unknown command!\n")
	}
}

// StartServer starts a node
func StartServer(minerAddress string) {
	miningAddress = minerAddress

	ln, err := net.Listen(protocol, localNodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	//myBlockVersion, myBestHeight := GetLocalHeightAndVersion()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn)
	}
}

// GetLocalHeightAndVersion ...
func GetLocalHeightAndVersion() (int, int) {
	bc := NewBlockchain()
	return bc.GetBestHeight(), bc.GetVersion()
}

// MempoolAddTxs add txs into mempool
func MempoolAddTxs(tx *Transaction) {
	mp.Lock.Lock()
	defer mp.Lock.Unlock()

	if mp.mempool[hex.EncodeToString(tx.ID)] == nil {
		mp.mempool[hex.EncodeToString(tx.ID)] = tx
		mp.txNums++
	}

	fmt.Println("a new tx has been send to mempool")
}

// MempoolDeleteTxs delete txs from mempool
func MempoolDeleteTxs(txs []*Transaction) {
	mp.Lock.Lock()
	defer mp.Lock.Unlock()

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		if mp.mempool[txID] != nil {
			delete(mp.mempool, txID)
			mp.txNums--
		}
		fmt.Println("a tx was deleted from mempool")
	}
}

// MempoolGetTxNums returns the number of txs in the mempool
func MempoolGetTxNums() int {
	mp.Lock.Lock()
	defer mp.Lock.Unlock()

	txNums := mp.txNums
	return txNums
}

// MempoolVerifyTxs returns Transactions which is verified
func MempoolVerifyTxs(bc *Blockchain) []*Transaction {
	mp.Lock.Lock()
	defer mp.Lock.Unlock()

	var txs []*Transaction
	for id := range mp.mempool {
		tx := mp.mempool[id]
		if bc.VerifyTransaction(tx) {
			txs = append(txs, tx)
		} else {
			delete(mp.mempool, id)
			mp.txNums--
		}
	}

	return txs
}
