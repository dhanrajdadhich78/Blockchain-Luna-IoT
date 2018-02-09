package app

import (
	"fmt"
	"log"
	"strconv"
	"time"
	b "wizeBlockchain/blockchain"
	w "wizeBlockchain/wallet"
)

func (cli *CLI) printChain(nodeID string) {
	bc := b.NewBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		fmt.Printf("Created at: %s\n", time.Unix(block.Timestamp, 0))
		pow := b.NewProofOfWork(block)
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

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := w.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
	fmt.Println("  startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println()
	fmt.Println("Exploring cmds:")
	//fmt.Println("  generateKey - generate KeyPair for exploring")
	fmt.Println("  generatePrivKey - generate KeyPair for exploring")
	fmt.Println("  getPubKey -privKey PRIKEY - generate PubKey from privateKey")
	fmt.Println("  getAddress -pubKey PUBKEY - convert pubKey to address")
	fmt.Println("  getPubKeyHash -address Address - get pubKeyHash of an address")
	fmt.Println("  validateAddress -addr Address - validate an address")
	fmt.Println("  getBlock -hash BlockHash - get a block with BlockHash")
}

func (cli *CLI) printBlock(blockHash, nodeID string) {
	bc := b.NewBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		hash := fmt.Sprintf("%x", block.Hash)
		if hash == blockHash {
			fmt.Printf("============ Block %x ============\n", block.Hash)
			fmt.Printf("Height: %d\n", block.Height)
			fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
			fmt.Printf("Created at : %s\n", time.Unix(block.Timestamp, 0))
			pow := b.NewProofOfWork(block)
			fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
			for _, tx := range block.Transactions {
				fmt.Println(tx)
			}
			fmt.Printf("\n\n")
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
