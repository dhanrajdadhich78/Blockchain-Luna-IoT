package blockchain

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"

	ecdsa "wizeBlock/wizeNode/crypto"
	"wizeBlock/wizeNode/utils"
)

const Version = byte(0x00)
const addressChecksumLen = 4

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{*private, public}

	return &wallet
}

// CreateWallet from private key
func CreateWallet(privateKey []byte) (*Wallet, error) {
	private, err := ecdsa.GetPrivateKey(nil, privateKey)
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
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{Version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := utils.Base58Encode(fullPayload)

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

////

// NewKeyPair
func NewKeyPair() (*ecdsa.PrivateKey, []byte) {
	// TODO: should we realize curve?
	//curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(nil, rand.Reader)
	if err != nil {
		fmt.Printf("Cant generate keys: %s", err)
		return nil, nil
	}
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return privKey, pubKey
}

// GetAddress
func GetAddress(pubKey []byte) []byte {
	pubKeyHash := HashPubKey(pubKey)

	versionedPayload := append([]byte{Version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := utils.Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	fullPayload := utils.Base58Decode([]byte(address))
	actualChecksum := fullPayload[len(fullPayload)-addressChecksumLen:]
	version := fullPayload[0]
	pubKeyHash := fullPayload[1 : len(fullPayload)-addressChecksumLen]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func GetPubKeyHash(address string) []byte {
	fullPayload := utils.Base58Decode([]byte(address))
	pubKeyHash := fullPayload[1 : len(fullPayload)-addressChecksumLen]
	fmt.Printf("PubKeyHash: %x\n", pubKeyHash)
	return pubKeyHash
}

// Checksum generates a Checksum for a public key
func Checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func GetAddresses(bc *Blockchain) []string {
	return bc.GetAddresses()
}

func GetWalletBalance(address string, bc *Blockchain) int {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	UTXOSet := UTXOSet{bc}
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}
