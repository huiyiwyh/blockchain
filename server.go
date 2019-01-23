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

type block struct {
	PeerFrom string
	Block    []byte
}

type getblocks struct {
	PeerFrom string
}

type getdata struct {
	PeerFrom string
	Type     string
	ID       []byte
}

type inv struct {
	PeerFrom string
	Type     string
	Items    [][]byte
}

type tx struct {
	PeerFrom    string
	Transaction []byte
}

type version struct {
	PeerFrom   string
	Version    int
	BestHeight int64
}

// LocalPeer defines local ip + port
var LocalPeer string

// blocksInTransit stores blocks hash
var blocksInTransit = [][]byte{}

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

func sendBlock(PeerTo string, b *Block) {
	data := block{LocalPeer, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(PeerTo, request)
}

func sendInv(PeerTo, kind string, items [][]byte) {
	inventory := inv{LocalPeer, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(PeerTo, request)
}

func sendGetBlocks(PeerTo string) {
	payload := gobEncode(getblocks{LocalPeer})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(PeerTo, request)
}

func sendGetData(PeerTo, kind string, id []byte) {
	payload := gobEncode(getdata{LocalPeer, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(PeerTo, request)
}

func sendTx(PeerTo string, tnx *Transaction) {
	data := tx{LocalPeer, tnx.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(PeerTo, request)
}

func sendVersion(PeerTo string, blockVersion int, bestHeight int64) {
	payload := gobEncode(version{LocalPeer, blockVersion, bestHeight})
	request := append(commandToBytes("version"), payload...)

	sendData(PeerTo, request)
}

func sendData(PeerTo string, data []byte) {
	conn, err := net.Dial(protocol, PeerTo)
	if err != nil {
		fmt.Printf("%s is not available\n", PeerTo)

		SToPGetPI <- &Notification{}
		pi := <-PToSPI
		knownPeers := MapToSlice(pi.Peers)

		var updatedPeers []string

		for _, Peer := range knownPeers {
			if Peer != PeerTo {
				updatedPeers = append(updatedPeers, Peer)
			}
		}

		knownPeers = updatedPeers
		SToPPeer <- knownPeers

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

	fmt.Printf("Recevied a new block! from %s\n", payload.PeerFrom)

	SToBBlock <- block

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.PeerFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	}

	go MindOtherPeerBlock(payload.PeerFrom, block)
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

	fmt.Printf("Recevied inventory with %d %s from %s\n", len(payload.Items), payload.Type, payload.PeerFrom)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.PeerFrom, "block", blockHash)

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

		SToMGetTxByHash <- &TxByHash{payload.PeerFrom, txID, nil}

		if txByHash := <-MToSTxByHash; txByHash.Tx == nil {
			sendGetData(txByHash.PeerFrom, "tx", payload.Items[0])
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

	fmt.Printf("Received getblocks command from %s\n", payload.PeerFrom)

	SToBGetBlocksHash <- &BlocksHash{payload.PeerFrom, nil}
	blockByHash := <-BToSBlocksHash

	sendInv(blockByHash.PeerFrom, "block", blockByHash.Hashs)
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

	fmt.Printf("Received getdata command from %s\n", payload.PeerFrom)

	if payload.Type == "block" {
		SToBGetBlockByHash <- &BlockByHash{payload.PeerFrom, payload.ID, nil}

		if blockByHash := <-BToSBlockByHash; blockByHash.Block != nil {
			sendBlock(blockByHash.PeerFrom, blockByHash.Block)
		}
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)

		SToMGetTxByHash <- &TxByHash{payload.PeerFrom, txID, nil}

		if txByHash := <-MToSTxByHash; txByHash.Tx != nil {
			sendTx(txByHash.PeerFrom, txByHash.Tx)
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

	fmt.Printf("Received tx command from %s\n", payload.PeerFrom)

	SToBGetBI <- &Notification{}
	blockchainInfo := <-BToSBI

	if blockchainInfo.IsMiner {
		SToMTx <- tx
	}

	go MindOtherPeerTx(payload.PeerFrom, tx)
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

	fmt.Printf("Received version command from %s\n", payload.PeerFrom)

	SToBGetBI <- &Notification{}
	nbcm := <-BToSBI

	myBestHeight := nbcm.Height
	myBlockVersion := nbcm.BlockHeader.Version

	foreignerBestHeight := payload.BestHeight

	if myBestHeight > foreignerBestHeight {
		sendVersion(payload.PeerFrom, myBlockVersion, myBestHeight)
	}

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.PeerFrom)
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

// StartServer starts a Peer
func StartServer(minerAddress string) {
	ln, err := net.Listen(protocol, LocalPeer)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bcm := NewBlockchainManager(minerAddress)
	go bcm.Processor()

	mm := NewMempoolManager()
	go mm.Processor()

	pm := NewPeerManager()
	go pm.Processor()

	knowPeers := []string{"191.167.1.111:3000", "191.167.1.146:3000"}
	SToPPeer <- knowPeers

	go MindOtherPeerVersion()

	go BoardcastMinedBlock()

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
	SToPGetPI <- &Notification{}
	npmi := <-PToSPI

	return MapToSlice(npmi.Peers)
}

// MindOtherPeerVersion ...
func MindOtherPeerVersion() {
	knowPeers := GetPeers()

	SToBGetBI <- &Notification{}
	nbcm := <-BToSBI

	myBestHeight := nbcm.Height
	myBlockVersion := nbcm.BlockHeader.Version

	for _, peer := range knowPeers {
		if peer != LocalPeer {
			sendVersion(peer, myBlockVersion, myBestHeight)
		}
	}
}

// MindOtherPeerTx ...
func MindOtherPeerTx(peerFrom string, tx *Transaction) {
	knowPeers := GetPeers()

	for _, peer := range knowPeers {
		if peer != LocalPeer && peer != peerFrom {
			sendInv(peer, "tx", [][]byte{tx.ID})
		}
	}
}

// MindOtherPeerBlock ...
func MindOtherPeerBlock(peerFrom string, block *Block) {
	knownPeers := GetPeers()

	for _, peer := range knownPeers {
		if peer != LocalPeer && peer != peerFrom {
			sendInv(peer, "block", [][]byte{block.Hash})
		}
	}
}

func BoardcastMinedBlock() {
	for {
		select {
		case minedblock := <-BToSBlock:
			knownPeers := GetPeers()

			for _, peer := range knownPeers {
				if peer != LocalPeer {
					sendInv(peer, "block", [][]byte{minedblock.Hash})
				}
			}
		default:
		}
	}
}
