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
)

var localNodeAddress string
var miningAddress string

var blocksInTransit = [][]byte{}

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

type version struct {
	NodeFrom   string
	Version    int
	BestHeight int64
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

func sendBlock(nodeTo string, b *Block) {
	data := block{localNodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(nodeTo, request)
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

func sendVersion(nodeTo string, blockVersion int, bestHeight int64) {
	payload := gobEncode(version{localNodeAddress, blockVersion, bestHeight})
	request := append(commandToBytes("version"), payload...)

	sendData(nodeTo, request)
}

func sendData(nodeTo string, data []byte) {
	conn, err := net.Dial(protocol, nodeTo)
	if err != nil {
		fmt.Printf("%s is not available\n", nodeTo)

		SToPGetPM <- &Notification{}
		npm := <-PToSSendPM
		knownNodes := MapToSlice(npm.Peers)

		var updatedNodes []string

		for _, node := range knownNodes {
			if node != nodeTo {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes
		SToPPeer <- knownNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
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

	fmt.Printf("Recevied a new block! from %s\n", payload.NodeFrom)

	SToBCMBlock <- block

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.NodeFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	}

	knownNodes := GetPeers()
	for _, node := range knownNodes {
		sendInv(node, "block", [][]byte{block.Hash})
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
		txID := hex.EncodeToString(payload.Items[0])

		SToMGetTxByHash <- &TxByHash{payload.NodeFrom, txID, nil}

		if txByHash := <-MToSSendTxByHash; txByHash.Tx == nil {
			sendGetData(txByHash.NodeFrom, "tx", payload.Items[0])
		}
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

	SToBCMGetBlocksHash <- &BlocksHash{payload.NodeFrom, nil}
	blockByHash := <-BCMToSBlocksHash

	sendInv(blockByHash.NodeFrom, "block", blockByHash.Hashs)
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

	if payload.Type == "block" {
		SToBCMGetBlockByHash <- &BlockByHash{payload.NodeFrom, payload.ID, nil}

		if blockByHash := <-BCMToSBlockByHash; blockByHash.Block != nil {
			sendBlock(blockByHash.NodeFrom, blockByHash.Block)
		}
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)

		SToMGetTxByHash <- &TxByHash{payload.NodeFrom, txID, nil}

		if txByHash := <-MToSSendTxByHash; txByHash.Tx != nil {
			sendTx(txByHash.NodeFrom, txByHash.Tx)
		}
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

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)

	fmt.Printf("Received tx command from %s\n", payload.NodeFrom)

	SToMTx <- tx

	knownNodes := GetPeers()
	for _, node := range knownNodes {
		sendInv(node, "tx", [][]byte{tx.ID})
	}
}

func handleVersion(request []byte) {
	var buff bytes.Buffer
	var payload version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received version command from %s\n", payload.NodeFrom)

	SToBCMGetBCM <- &Notification{}
	nbcm := <-BCMToSSendBCM

	myBestHeight := nbcm.Height
	myBlockVersion := nbcm.BlockHeader.Version

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

	bcm := NewBlockchainManager()
	go bcm.Processor()

	mp := NewMempoolManager()
	go mp.Processor()

	pm := NewPeerManager()
	go pm.Processor()

	knowNodes := []string{"191.167.2.1:3000", "191.167.2.17:3000", "191.167.1.111:3000", "191.167.1.146:3000"}
	SToPPeer <- knowNodes

	SToBCMGetBCM <- &Notification{}
	nbcm := <-BCMToSSendBCM

	myBestHeight := nbcm.Height
	myBlockVersion := nbcm.BlockHeader.Version

	for _, node := range knowNodes {
		sendVersion(node, myBlockVersion, myBestHeight)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn)
	}
}

// GetPeers ...
func GetPeers() []string {
	SToPGetPM <- &Notification{}
	npm := <-PToSSendPM

	return MapToSlice(npm.Peers)
}
