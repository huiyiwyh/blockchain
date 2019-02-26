package main

import (
	"flag"
	"fmt"
	"log"

	"os"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a Blockchain and send genesis block reward to address")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of address")
	fmt.Println("  getwalletinfo - gets all addresses from the wallet file")
	fmt.Println("  getBlockchaininfo - Print all the blocks of the blockchain")
	fmt.Println("  getpeerinfo - Print all the nodes of the currentnode")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  createtx -from FROM -to TO -amount AMOUNT -mine - CreateTx AMOUNT of coins from from address to to. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node -miner enables mining with a wallet address")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	deleteWalletCmd := flag.NewFlagSet("deletewallet", flag.ExitOnError)
	getWalletInfoCmd := flag.NewFlagSet("getwalletinfo", flag.ExitOnError)
	getPeerInfoCmd := flag.NewFlagSet("getpeerinfo", flag.ExitOnError)
	getBlockchainInfoCmd := flag.NewFlagSet("getBlockchaininfo", flag.ExitOnError)
	createTxCmd := flag.NewFlagSet("createtx", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	createBlockchainAddress := createBlockchainCmd.String("address", "", "")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	deleteWalletAddress := deleteWalletCmd.String("address", "", "delete wallet address")
	createTxFrom := createTxCmd.String("from", "", "Source wallet address")
	createTxTo := createTxCmd.String("to", "", "Destination wallet address")
	createTxAmount := createTxCmd.Int("amount", 0, "Amount to send")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "deletewallet":
		err := deleteWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getwalletinfo":
		err := getWalletInfoCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getpeerInfoCmd":
		err := getPeerInfoCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getblockchaininfo":
		err := getBlockchainInfoCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createtx":
		err := createTxCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getpeerinfo":
		err := getPeerInfoCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if deleteWalletCmd.Parsed() {
		if *deleteWalletAddress == "" {
			deleteWalletCmd.Usage()
			os.Exit(1)
		}
		cli.deleteWallet(*deleteWalletAddress)
	}

	if getWalletInfoCmd.Parsed() {
		cli.getWalletsInfo()
	}

	if getPeerInfoCmd.Parsed() {
		cli.getPeerInfo()
	}

	if getBlockchainInfoCmd.Parsed() {
		cli.getblockchaininfo()
	}

	if createTxCmd.Parsed() {
		if *createTxFrom == "" || *createTxTo == "" || *createTxAmount <= 0 {
			createTxCmd.Usage()
			os.Exit(1)
		}

		cli.createTx(*createTxFrom, *createTxTo, *createTxAmount)
	}

	if startNodeCmd.Parsed() {
		cli.startNode(*startNodeMiner)
	}
}
