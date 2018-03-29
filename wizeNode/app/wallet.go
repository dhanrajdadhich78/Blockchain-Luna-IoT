package app

import (
	"log"

	blockchain "wizeBlock/wizeNode/blockchain"
	"wizeBlock/wizeNode/utils"
)

// TODO-34
func GetWalletCredits(address string, nodeID string, bc *blockchain.Blockchain) int {
	if !blockchain.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	UTXOSet := blockchain.UTXOSet{bc}

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}
