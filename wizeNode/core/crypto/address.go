package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const Version = byte(0x00)
const addressChecksumLen = 4

// NewKeyPair
func NewKeyPair() (*PrivateKey, []byte) {
	// TODO: should we realize curve?
	//curve := elliptic.P256()
	privKey, err := GenerateKey(nil, rand.Reader)
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
	return GetAddressFromPubKeyHash(pubKeyHash)
}

func GetAddressFromPubKeyHash(pubKeyHash []byte) []byte {
	versionedPayload := append([]byte{Version}, pubKeyHash...)
	checksum := Checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)
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
	fullPayload := Base58Decode([]byte(address))
	actualChecksum := fullPayload[len(fullPayload)-addressChecksumLen:]

	version := fullPayload[0]
	pubKeyHash := fullPayload[1 : len(fullPayload)-addressChecksumLen]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

func GetPubKeyHash(address string) []byte {
	fullPayload := Base58Decode([]byte(address))
	pubKeyHash := fullPayload[1 : len(fullPayload)-addressChecksumLen]
	return pubKeyHash
}

// Checksum generates a Checksum for a public key
func Checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}
