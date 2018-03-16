package app

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	blockchain "wizeBlock/wizeNode/blockchain"
	"wizeBlock/wizeNode/utils"
)

func (cli *CLI) generatePrivKey() {
	private, _ := blockchain.NewKeyPair()
	fmt.Println(hex.EncodeToString(private.D.Bytes()))

}

func (cli *CLI) getAddress(pubKey string) {
	public, _ := hex.DecodeString(pubKey)

	pubKeyHash := blockchain.HashPubKey(public)

	versionedPayload := append([]byte{blockchain.Version}, pubKeyHash...)
	fullPayload := append(versionedPayload, blockchain.Checksum(versionedPayload)...)

	fmt.Println()
	fmt.Printf("PubKey     : %s\n", pubKey)
	fmt.Printf("PubKeyHash : %x\n", pubKeyHash)
	fmt.Printf("Address    : %s\n", utils.Base58Encode(fullPayload))
}

//func (cli *CLI) getAddress(pubKey string) {
//	public, _ := hex.DecodeString(pubKey)
//
//	pubKeyHash := w.HashPubKey(public)
//
//	versionedPayload := append([]byte{w.Version}, pubKeyHash...)
//	checksum := w.Checksum(versionedPayload)
//
//	fullPayload := append(versionedPayload, checksum...)
//	address := s.Base58Encode(fullPayload)
//	fmt.Println()
//	//fmt.Printf("PubKey : %s\nAddress: %s\n", pubKey, address)
//	fmt.Printf("PubKey     : %s\n", pubKey)
//	fmt.Printf("PubKeyHash : %x\n", pubKeyHash)
//	fmt.Printf("Address    : %s\n", address)
//}

func (cli *CLI) getPubKey(privateKey string) {
	curve := elliptic.P256()
	priv_key, _ := hex.DecodeString(privateKey)
	x, y := curve.ScalarBaseMult(priv_key)
	pubKey := append(x.Bytes(), y.Bytes()...)
	fmt.Println(hex.EncodeToString(pubKey))
}

func (cli *CLI) getPubKeyHash(address string) {
	pubKeyHash := utils.Base58Decode([]byte(address))
	fmt.Printf("%x\n", pubKeyHash[1:len(pubKeyHash)-4])
}

func (cli *CLI) validateAddr(address string) {
	fmt.Printf("Address: %s\n", address)
	if !blockchain.ValidateAddress(address) {
		fmt.Println("Not valid!")
	} else {
		fmt.Println("Valid!")
	}
}

// print

func (cli *CLI) printBlock(blockHash, nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
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
			pow := blockchain.NewProofOfWork(block)
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
