package app

import (
	"log"

	b "wizeBlock/wizeNode/blockchain"
	s "wizeBlock/wizeNode/services"
	w "wizeBlock/wizeNode/wallet"
)

func GetWalletCredits(address string, nodeID string, bc *b.Blockchain) int {
	if !w.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	UTXOSet := b.UTXOSet{bc}

	balance := 0
	pubKeyHash := s.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}
