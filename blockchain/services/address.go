package services

import (
"github.com/coin-network/curve"
"crypto/ecdsa"
"fmt"
)

type PublicKey ecdsa.PublicKey
type PrivateKey ecdsa.PrivateKey

func MakeAddress() {

	KoblitzCurve := curve.S256() // see https://godoc.org/github.com/btcsuite/btcd/btcec#S256

	privkey, err := curve.NewPrivateKey(KoblitzCurve)

	if err != nil {
		panic("Error")
	}

	fmt.Println(" Private Key ")
	fmt.Println(privkey.D)

	fmt.Println(" Public Key ")
	pubkey := (privkey.PublicKey)
	fmt.Println(pubkey.X)
	fmt.Println(pubkey.Y)

	fmt.Println("-------")
	fmt.Println(" New Address ")
	address := privkey.PubKey().ToAddress()
	fmt.Println(address)
}
