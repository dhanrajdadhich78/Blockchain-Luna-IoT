package app

import (
	"log"

	w "wizeBlockchain/wallet"
	s "wizeBlockchain/services"
	b "wizeBlockchain/blockchain"
	"fmt"
)

func GetWalletCredits(address string, nodeID string) int {
	if !w.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	fmt.Println(nodeID)
	bc := b.NewBlockchain(nodeID)
	fmt.Println(bc)
	UTXOSet := b.UTXOSet{bc}
	fmt.Println(UTXOSet)
	//
	fmt.Println(nodeID)
	balance := 0
	pubKeyHash := s.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	fmt.Println(nodeID)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Println(nodeID)
	return balance
}
