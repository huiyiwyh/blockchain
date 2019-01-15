package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const nodesdbFile = "nodesinfo.db"
const nodesInfoBucket = "nodesinfo"
const centralAddress = "191.167.1.146:3000"

type NodeInfo struct {
	Wallets map[string]string
	Nodes   map[string]string
}

type NodesInfo map[string]*NodeInfo

func CreateOrLoadNodesInfodb() {
	db, err := bolt.Open(nodesdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
}

func CreateNodesInfo() {
	db, err := bolt.Open(nodesdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		if nodesinfoBucket := tx.Bucket([]byte(nodesInfoBucket)); nodesinfoBucket != nil {
			return nil
		}

		n, err := tx.CreateBucket([]byte(nodesInfoBucket))
		if err != nil {
			log.Panic(err)
		}

		nodeInfo := &NodeInfo{}
		nodeInfo.Nodes = make(map[string]string)
		nodeInfo.Wallets = make(map[string]string)
		nodeInfo.Nodes[localNodeAddress] = localNodeAddress
		nodeInfo.Nodes[centralAddress] = centralAddress

		if _, err := os.Stat(walletFile); !os.IsNotExist(err) {
			ws, err := NewWallets()
			if err != nil {
				log.Println(err)
			}
			walletsAddresses := ws.GetAddresses()

			for _, walletAddress := range walletsAddresses {
				nodeInfo.Wallets[walletAddress] = walletAddress
			}
		}

		dataBytes := SerializeNodeInfo(nodeInfo)

		err = n.Put([]byte(localNodeAddress), []byte(dataBytes))
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func NewNodesInfo() NodesInfo {
	newNodesInfo := make(map[string]*NodeInfo)

	db, err := bolt.Open(nodesdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		n := tx.Bucket([]byte(nodesInfoBucket))

		c := n.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			node := string(k)
			nodeInfo := DeserializeNodeInfo(v)
			newNodesInfo[node] = nodeInfo
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return newNodesInfo
}

func UpdateNodesInfo(nodesInfo NodesInfo) {
	db, err := bolt.Open(nodesdbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		n := tx.Bucket([]byte(nodesInfoBucket))

		for nodeAddress, nodeInfo := range nodesInfo {
			nodeInfoBytes := SerializeNodeInfo(nodeInfo)

			err := n.Put([]byte(nodeAddress), []byte(nodeInfoBytes))
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// Serialize serializes the NodeInfo
func SerializeNodeInfo(nodeInfo *NodeInfo) []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(nodeInfo)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeNodeInfo(data []byte) *NodeInfo {
	var nodeInfo NodeInfo

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&nodeInfo)
	if err != nil {
		log.Panic(err)
	}

	return &nodeInfo
}

func IsNodeInfoDifferent(foreignerNodeAddress string, foreignerNodeInfo *NodeInfo, localNodeInfo *NodeInfo) bool {
	for k, _ := range localNodeInfo.Nodes {
		if foreignerNodeInfo.Nodes[k] == "" {
			return true
		}
	}

	for k, _ := range localNodeInfo.Wallets {
		if foreignerNodeInfo.Wallets[k] == "" {
			return true
		}
	}

	for k, _ := range foreignerNodeInfo.Nodes {
		if localNodeInfo.Nodes[k] == "" {
			return true
		}
	}

	for k, _ := range foreignerNodeInfo.Wallets {
		if localNodeInfo.Wallets[k] == "" {
			return true
		}
	}

	return false
}
