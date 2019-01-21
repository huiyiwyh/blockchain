package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func getbcm(w http.ResponseWriter, r *http.Request) {
	CToBCMGetBCM <- &Notification{}
	nbcm := <-BCMToCSendBCM

	w.Write(nbcm.Hash)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// Processor ...
func (c *CLI) Processor() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getbcm", getbcm)

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
	// resp, err := http.Post("127.0.0.1:15927/",
	// 	"application/x-www-form-urlencoded",
	// 	strings.NewReader(""))
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	// handle error
	// }

	// fmt.Println(string(body))
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
