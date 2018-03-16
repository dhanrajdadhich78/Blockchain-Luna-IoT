package app

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	blockchain "wizeBlock/wizeNode/blockchain"
)

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println()
	fmt.Println("Exploring cmds:")
	fmt.Println("  generatePrivKey - generate KeyPair for exploring")
	fmt.Println("  getPubKey -privKey PRIKEY - generate PubKey from privateKey")
	fmt.Println("  getPubKeyHash -address Address - get pubKeyHash of an address")
	fmt.Println("  getAddress -pubKey PUBKEY - convert pubKey to address")
	fmt.Println("  validateAddress -addr Address - validate an address")
	fmt.Println("  getBlock -hash BlockHash - get a block with BlockHash")
}

func (cli *CLI) createBlockchain(address string, nodeID string) {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := blockchain.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	wallet := wallets.GetWallet(address)

	fmt.Printf("Your new address: %s\n", address)
	fmt.Println("Private key: ", hex.EncodeToString(wallet.GetPrivateKey(wallet)))
	fmt.Println("Public key: ", hex.EncodeToString(wallet.GetPublicKey(wallet)))
}

func (cli *CLI) getBalance(address string, nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	balance := GetWalletCredits(address, nodeID, bc)

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := blockchain.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) printChain(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		fmt.Printf("Created at: %s\n", time.Unix(block.Timestamp, 0))
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !blockchain.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !blockchain.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := blockchain.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	if wallet == nil {
		fmt.Println("The Address doesn't belongs to you!")
		return
	}
	tx := blockchain.NewUTXOTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.NewCoinbaseTX(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		SendTx(knownNodes[0], tx) //TODO: проверять остаток на балансе с учетом незамайненых транзакций, во избежание двойного использования выходов
	}

	fmt.Println("Success!")
}

func (cli *CLI) startNode(nodeID, minerAddress string, apiAddr string) {
	fmt.Printf("Starting node %s\n", nodeID)
	nodeADD := os.Getenv("NODE_ADD")
	if len(minerAddress) > 0 {
		if blockchain.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
			//StartServer(nodeID, minerAddress, apiAddr)
			node := NewNode(nodeID)
			node.apiAddr = apiAddr
			node.nodeID = nodeID
			node.nodeADD = nodeADD
			node.Run(minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	//StartServer(nodeID, minerAddress, apiAddr)

	node := NewNode(nodeID)
	node.apiAddr = apiAddr
	node.nodeID = nodeID
	node.nodeADD = nodeADD
	node.Run(minerAddress)
}
