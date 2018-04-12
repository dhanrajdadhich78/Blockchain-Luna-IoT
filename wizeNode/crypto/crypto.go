package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"math/big"

	"github.com/btccom/secp256k1-go/secp256k1"
)

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

type PrivateKey struct {
	PublicKey
	D *big.Int
}

func rand32() [32]byte {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		panic(err)
	}
	return key
}

func GenerateKey(c elliptic.Curve, rand io.Reader) (*PrivateKey, error) {
	// context
	params := uint(secp256k1.ContextSign | secp256k1.ContextVerify)
	ctx, err := secp256k1.ContextCreate(params)
	if err != nil {
		return nil, err
	}
	log.Printf("%+v\n", ctx)

	// generate private key
	privateKey := rand32()
	log.Printf("Private Key: %s\n", hex.EncodeToString(privateKey[:]))
	res := secp256k1.ContextRandomize(ctx, privateKey)
	log.Printf("Result of randomize: %d\n", res)

	// verify private key
	_, err = secp256k1.EcSeckeyVerify(ctx, privateKey[:])
	if err != nil {
		return nil, err
	}
	log.Printf("Verifying was successful\n")

	// get the public key
	_, publicKeyStruct, err := secp256k1.EcPubkeyCreate(ctx, privateKey[:])
	if err != nil {
		return nil, err
	}
	log.Printf("Public Key: %+v\n", publicKeyStruct)

	_, publicKey, err := secp256k1.EcPubkeySerialize(ctx, publicKeyStruct, secp256k1.EcUncompressed)
	if err != nil {
		return nil, err
	}
	// publicKey has 65 bytes when Uncompressesed
	// first byte equals 0x04
	log.Printf("Public Key: %+v\n", hex.EncodeToString(publicKey[:]))
	log.Printf("Public Key X: %+v\n", hex.EncodeToString(publicKey[1:33]))
	log.Printf("Public Key Y: %+v\n", hex.EncodeToString(publicKey[33:]))

	priv := new(PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = new(big.Int).SetBytes(privateKey[:])
	priv.PublicKey.X = new(big.Int).SetBytes(publicKey[1:33])
	priv.PublicKey.Y = new(big.Int).SetBytes(publicKey[33:])

	secp256k1.ContextDestroy(ctx)

	return priv, nil
}
