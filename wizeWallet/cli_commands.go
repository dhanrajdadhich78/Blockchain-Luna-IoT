package main

import (
	"encoding/hex"
	"fmt"
	"log"
	//"os"
	//"strconv"
	//"time"

	//"wizeBlock/wizeNode/blockchain"
	//"wizeBlock/wizeNode/crypto"
	"wizeBlock/wizeNode/wallet"
)

// TODO: add REST API Client to wizeNode REST API Service

// DONE: createwallet
// DONE: listaddresses
// TODO: getwallet
// TODO: getbalance
// TODO: send (prepare/sign)
// TODO: fix nodeID

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet - Generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - Lists all addresses from the wallet file")
	fmt.Println("  getwallet -address ADDRESS - Get wallet info")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  getBlock -hash BlockHash - get a block with BlockHash")
	fmt.Println("  reindexutxo - Rebuilds the UTXO set")
}

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := wallet.NewWalletsExt("wallet%s.data", nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)
	walletNew := wallets.GetWallet(address)

	fmt.Printf("Your new address: %s\n", address)
	fmt.Println("Private key: ", hex.EncodeToString(walletNew.GetPrivateKey()))
	fmt.Println("Public key:  ", hex.EncodeToString(walletNew.GetPublicKey()))
}

func (cli *CLI) listAddresses(nodeID string) {
	var addresses []string = []string{}
	wallets, err := wallet.NewWalletsExt("wallet%s.data", nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses = wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) getWallet(address string, nodeID string) {
}

func (cli *CLI) getBalance(address string, nodeID string) {
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
}

func (cli *CLI) printChain(nodeID string) {
}

func (cli *CLI) printBlock(blockHash, nodeID string) {
}

func (cli *CLI) reindexUTXO(nodeID string) {
}
