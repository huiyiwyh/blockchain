package main

import (
	"net/http"
	"time"
)

func getBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	CToBGetBI <- &Notification{}
	data := gobEncode(<-BToCBI)
	w.Write(data)
}

func getMempoolInfo(w http.ResponseWriter, r *http.Request) {
	CToMGetMI <- &Notification{}
	data := gobEncode(<-MToCMI)
	w.Write(data)

}

func getPeerInfo(w http.ResponseWriter, r *http.Request) {
	CToPGetPI <- &Notification{}
	data := gobEncode(<-PToCPI)
	w.Write(data)
}

func addPeer(w http.ResponseWriter, r *http.Request) {

}

func deletePeer(w http.ResponseWriter, r *http.Request) {

}

// Processor ...
func (c *CLI) Processor() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getBlockchainMangerInfo", getBlockchainInfo)
	mux.HandleFunc("/getMempoolInfo", getMempoolInfo)
	mux.HandleFunc("/getPeerInfo", getPeerInfo)
	mux.HandleFunc("/addPeer", addPeer)
	mux.HandleFunc("/deletePeer", deletePeer)

	server := http.Server{
		Addr:           "0.0.0.0:8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
