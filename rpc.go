package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// func getbalanceinfo(w http.ResponseWriter, r *http.Request) {
// var m msgGetBalance

// decoder := json.NewDecoder(r.Body)
// if err := decoder.Decode(&m); err != nil {
// 	respondWithJSON(w, r, http.StatusBadRequest, r.Body)
// 	return
// }

// if !ValidateAddress(m.Address) {
// 	log.Panic("ERROR: Address is not valid")
// }
// bc := NewBlockchain()

// UTXOSet := UTXOSet{bc.tip}

// balance := 0
// pubKeyHash := Base58Decode([]byte(m.Address))
// pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
// UTXOs := UTXOSet.FindUTXOByHash(pubKeyHash)

// for _, out := range UTXOs {
// 	balance += out.Value
// }

// fmt.Printf("Balance of '%s': %d\n", m.Address, balance)
// }

// func getwalletsinfo(w http.ResponseWriter, r *http.Request) {
// var m msgGetWalletsInfo

// decoder := json.NewDecoder(r.Body)
// if err := decoder.Decode(&m); err != nil {
// 	respondWithJSON(w, r, http.StatusBadRequest, r.Body)
// 	return
// }

// wallets, err := NewWallets()
// if err != nil {
// 	log.Panic(err)
// }
// addresses := wallets.GetAddresses()

// for _, address := range addresses {
// 	fmt.Println(address)
// }
// }

// func getBlockchaininfo(w http.ResponseWriter, r *http.Request) {
// var m msgGetBlockchainInfo

// decoder := json.NewDecoder(r.Body)
// if err := decoder.Decode(&m); err != nil {
// 	respondWithJSON(w, r, http.StatusBadRequest, r.Body)
// 	return
// }

// bc := NewBlockchain()
// bci := bc.Iterator(bc.tip)

// for {
// 	block := bci.Next()

// 	fmt.Printf("============ Block %x ============\n", block.Hash)
// 	fmt.Printf("Height: %d\n", block.Height)
// 	fmt.Printf("Prev. block: %x\n", block.BlockHeader.PrevBlockHash)
// 	pow := NewProofOfWork(block)
// 	fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
// 	for _, tx := range block.Transactions {
// 		fmt.Println(tx)
// 	}
// 	fmt.Printf("\n\n")

// 	if block.BlockHeader.PrevBlockHash == nil {
// 		break
// 	}
// }
// }

// func createTx(w http.ResponseWriter, r *http.Request) {
// var m msgSendTx

// decoder := json.NewDecoder(r.Body)
// if err := decoder.Decode(&m); err != nil {
// 	respondWithJSON(w, r, http.StatusBadRequest, r.Body)
// 	return
// }

// from := string(m.From)
// to := string(m.To)
// amount := m.Amount

// if !ValidateAddress(from) {
// 	log.Panic("ERROR: Sender address is not valid")
// }
// if !ValidateAddress(to) {
// 	log.Panic("ERROR: Recipient address is not valid")
// }

// bc := NewBlockchain()
// UTXOSet := UTXOSet{bc.tip}

// wallets, err := NewWallets()
// if err != nil {
// 	log.Panic(err)
// }
// wallet := wallets.GetWallet(from)

// tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

// // boardcast
// sendTx("", tx)
// fmt.Println("send to ")

// fmt.Printf("Success!\n")
// }

// func getBCM(w http.ResponseWriter, r *http.Request) {
// var m msgGetBCM

// decoder := json.NewDecoder(r.Body)
// if err := decoder.Decode(&m); err != nil {
// 	respondWithJSON(w, r, http.StatusBadRequest, r.Body)
// 	return
// }

// 	CToBCMGetBCM <- &Notification{}
// 	nbcm := <-BCMToCSendBCM

// 	w.Write(nbcm.Hash)
// }

// func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
// 	response, err := json.MarshalIndent(payload, "", "  ")
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("HTTP 500: Internal Server Error"))
// 		return
// 	}
// 	w.WriteHeader(code)
// 	w.Write(response)
// }

// Processor ...
func (c *CLI) Processor() {
	mux := http.NewServeMux()

	// mux.HandleFunc("/getbalanceinfo", getbalanceinfo)
	// mux.HandleFunc("/getwalletsinfo", getwalletsinfo)
	// mux.HandleFunc("/getBlockchaininfo", getBlockchaininfo)
	// mux.HandleFunc("/createTx", createTx)
	// mux.HandleFunc("/getbcm", getBCM)

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

func httpPost() {
	resp, err := http.Post("127.0.0.1:15927/",
		"application/x-www-form-urlencoded",
		strings.NewReader(""))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

func httpGet() {
	resp, err := http.Get("127.0.0.1:15927/")
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}
