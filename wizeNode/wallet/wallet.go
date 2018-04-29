package wallet

import (
	"fmt"

	"wizeBlock/wizeNode/crypto"
)

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey crypto.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := crypto.NewKeyPair()
	wallet := Wallet{*private, public}

	return &wallet
}

// CreateWallet from private key
func CreateWallet(privateKey []byte) (*Wallet, error) {
	private, err := crypto.GetPrivateKey(nil, privateKey)
	if err != nil {
		fmt.Printf("Cant generate keys: %s", err)
		return nil, err
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	wallet := Wallet{*private, public}

	return &wallet, nil
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := crypto.HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{crypto.Version}, pubKeyHash...)
	checksum := crypto.Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := crypto.Base58Encode(fullPayload)

	return address
}

func (w Wallet) GetPrivateKey() []byte {
	return w.PrivateKey.D.Bytes()
}
func (w Wallet) GetPublicKey() []byte {
	//public := append(w.PrivateKey.X.Bytes(), w.PrivateKey.Y.Bytes()...)
	public := w.PublicKey
	return public
}
