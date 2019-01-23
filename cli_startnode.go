package main

import (
	"fmt"
	"log"
	"os"
)

func (cli *CLI) startNode(minerAddress string) {
	fmt.Printf("Starting node\n")
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Println("Wrong miner address!")
			os.Exit(1)
		}
	}

	go cli.Processor()
	StartServer(minerAddress)
}
