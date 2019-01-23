package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
)

func (cli *CLI) getPeerInfo() {
	resp, err := http.Get("http://127.0.0.1:8080/getPeerInfo")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	var peerInfo PeerInfo

	dec := gob.NewDecoder(resp.Body)
	err = dec.Decode(&peerInfo)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("=============================")
	for _, peer := range peerInfo.Peers {
		fmt.Println(peer)
	}
}

func (cli *CLI) addPeer() {

}

func (cli *CLI) deletePeer() {

}
