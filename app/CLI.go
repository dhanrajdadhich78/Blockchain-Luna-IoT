package app

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	b "wizeBlockchain/blockchain"
	w "wizeBlockchain/wallet"
)

// CLI responsible for processing command line arguments
type CLI struct{}

func (cli *CLI) createBlockchain(address string, nodeID string) {
	if !w.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := b.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()
	UTXOSet := b.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := w.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	wallet := wallets.GetWallet(address)

	fmt.Printf("Your new address: %s\n", address)
	fmt.Println("Private key: ", hex.EncodeToString(wallet.GetPrivateKey(wallet)))
	fmt.Println("Public key: ", hex.EncodeToString(wallet.GetPublicKey(wallet)))

}
func (cli *CLI) getBalance(address string, nodeID string) {
	bc := b.NewBlockchain(nodeID)
	balance := GetWalletCredits(address, nodeID, bc)

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !w.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !w.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := b.NewBlockchain(nodeID)
	UTXOSet := b.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := w.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	if wallet == nil {
		fmt.Println("The Address doesn't belongs to you!")
		return
	}
	tx := b.NewUTXOTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := b.NewCoinbaseTX(from, "")
		txs := []*b.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		SendTx(knownNodes[0], tx) //TODO: проверять остаток на балансе с учетом незамайненых транзакций, во избежание двойного использования выходов
	}

	fmt.Println("Success!")
}

func (cli *CLI) startNode(nodeID, minerAddress string, apiAddr string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if w.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
			//StartServer(nodeID, minerAddress, apiAddr)
			node := NewNode(nodeID)
			node.apiAddr = apiAddr
			node.nodeID = nodeID
			node.Run(minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	//StartServer(nodeID, minerAddress, apiAddr)

	node := NewNode(nodeID)
	node.apiAddr = apiAddr
	node.nodeID = nodeID
	node.Run(minerAddress)
}

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := b.NewBlockchain(nodeID)
	UTXOSet := b.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

// Run parses command line arguments and processes commands
func (cli *CLI) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID must be set!")
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	//generateKeyCmd := flag.NewFlagSet("generateKey", flag.ExitOnError)
	generatePrivKeyCmd := flag.NewFlagSet("generatePrivKey", flag.ExitOnError)
	getPubKeyCmd := flag.NewFlagSet("getPubKey", flag.ExitOnError)
	getAddressCmd := flag.NewFlagSet("getAddress", flag.ExitOnError)
	getPubKeyHashCmd := flag.NewFlagSet("getPubKeyHash", flag.ExitOnError)
	validateAddrCmd := flag.NewFlagSet("validateAddress", flag.ExitOnError)
	getBlockCmd := flag.NewFlagSet("getBlock", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to Address")
	privateKey := getPubKeyCmd.String("privKey", "", "generate PubKey based on this")
	pubKey := getAddressCmd.String("pubKey", "", "the key where address generated")
	pubKeyAddress := getPubKeyHashCmd.String("address", "", "the pub address")
	address := validateAddrCmd.String("addr", "", "the public address")
	blockHash := getBlockCmd.String("hash", "", "the block hash")
	apiAddr := startNodeCmd.String("api", "4000", "Enable API server. Default port 4000")

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
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "validateAddress":
		err := validateAddrCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "generatePrivKey":
		err := generatePrivKeyCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getPubKey":
		err := getPubKeyCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getPubKeyHash":
		err := getPubKeyHashCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getAddress":
		err := getAddressCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBlock":
		err := getBlockCmd.Parse(os.Args[2:])
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
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress, nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}
	if reindexUTXOCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")

		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID, *startNodeMiner, *apiAddr)
	}

	if generatePrivKeyCmd.Parsed() {
		cli.generatePrivKey()
	}
	if getPubKeyCmd.Parsed() {
		if *privateKey == "" {
			getPubKeyCmd.Usage()
			os.Exit(1)
		}
		cli.getPubKey(*privateKey)
	}
	if getAddressCmd.Parsed() {
		if *pubKey == "" {
			getAddressCmd.Usage()
			os.Exit(1)
		}
		cli.getAddress(*pubKey)
	}

	if getPubKeyHashCmd.Parsed() {
		if *pubKeyAddress == "" {
			getPubKeyHashCmd.Usage()
			os.Exit(1)
		}

		cli.getPubKeyHash(*pubKeyAddress)
	}

	if validateAddrCmd.Parsed() {
		if *address == "" {
			validateAddrCmd.Usage()
			os.Exit(1)
		}
		cli.validateAddr(*address)
	}

	if getBlockCmd.Parsed() {
		if *blockHash == "" {
			getBlockCmd.Usage()
			os.Exit(1)
		}

		cli.printBlock(*blockHash, nodeID)
	}
}
