package crypto

import (
	"crypto/elliptic"
	"crypto/rand"
	//"encoding/hex"
	"io"
	"log"
	"math/big"

	"github.com/btccom/secp256k1-go/secp256k1"
)

// TODO: C library & cgo for different platforms
// TODO: working with contexts

// TODO: should we add curve?
type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

type PrivateKey struct {
	PublicKey
	D *big.Int
}

// TODO: more random?
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
	//log.Printf("%+v\n", ctx)

	// generate private key
	privateKey := rand32()
	//log.Printf("Private Key: %s\n", hex.EncodeToString(privateKey[:]))
	secp256k1.ContextRandomize(ctx, privateKey)
	//log.Printf("Result of randomize: %d\n", res)

	// verify private key
	_, err = secp256k1.EcSeckeyVerify(ctx, privateKey[:])
	if err != nil {
		return nil, err
	}
	//log.Printf("Verifying was successful\n")

	// get the public key
	_, publicKeyStruct, err := secp256k1.EcPubkeyCreate(ctx, privateKey[:])
	if err != nil {
		return nil, err
	}
	//log.Printf("Public Key: %+v\n", publicKeyStruct)

	_, publicKey, err := secp256k1.EcPubkeySerialize(ctx, publicKeyStruct, secp256k1.EcUncompressed)
	if err != nil {
		return nil, err
	}
	// publicKey has 65 bytes when Uncompressesed
	// first byte equals 0x04
	//log.Printf("Public Key:   %s\n", hex.EncodeToString(publicKey[:]))
	//log.Printf("Public Key X: %s\n", hex.EncodeToString(publicKey[1:33]))
	//log.Printf("Public Key Y: %s\n", hex.EncodeToString(publicKey[33:]))

	priv := new(PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = new(big.Int).SetBytes(privateKey[:])
	priv.PublicKey.X = new(big.Int).SetBytes(publicKey[1:33])
	priv.PublicKey.Y = new(big.Int).SetBytes(publicKey[33:])

	secp256k1.ContextDestroy(ctx)

	return priv, nil
}

func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	// context
	params := uint(secp256k1.ContextSign | secp256k1.ContextVerify)
	ctx, err := secp256k1.ContextCreate(params)
	if err != nil {
		return nil, nil, err
	}
	//log.Printf("%+v\n", ctx)

	publicKey := make([]byte, 1)
	publicKey[0] = 0x04
	publicKey = append(publicKey, priv.PublicKey.X.Bytes()...)
	publicKey = append(publicKey, priv.PublicKey.Y.Bytes()...)
	//log.Printf("Public Key: %s\n", hex.EncodeToString(publicKey))

	privateKey := priv.D.Bytes()
	_, ecdsaSignature, err := secp256k1.EcdsaSign(ctx, hash, privateKey)
	if err != nil {
		return nil, nil, err
	}
	//log.Printf("Signature: %+v\n", ecdsaSignature)

	//_, serializedDer, err := secp256k1.EcdsaSignatureSerializeDer(ctx, ecdsaSignature)
	//if err != nil {
	//	return nil, nil, err
	//}
	//log.Printf("Signature DER: %s\n", hex.EncodeToString(serializedDer[:]))

	_, serializedCompact, err := secp256k1.EcdsaSignatureSerializeCompact(ctx, ecdsaSignature)
	if err != nil {
		return nil, nil, err
	}
	//log.Printf("Signature Compact: %s\n", hex.EncodeToString(serializedCompact[:]))
	//log.Printf("Signature R: %s\n", hex.EncodeToString(serializedCompact[:32]))
	//log.Printf("Signature S: %s\n", hex.EncodeToString(serializedCompact[32:]))

	r = new(big.Int).SetBytes(serializedCompact[:32])
	s = new(big.Int).SetBytes(serializedCompact[32:])

	secp256k1.ContextDestroy(ctx)

	return r, s, nil
}

// TODO: To avoid accepting malleable signature, only ECDSA
// signatures in lower-S form are accepted. If you need to accept ECDSA
// sigantures from sources that do not obey this rule, apply
// EcdsaSignatureNormalize() prior to validation (however, this results in
// malleable signatures)
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	// context
	params := uint(secp256k1.ContextSign | secp256k1.ContextVerify)
	ctx, err := secp256k1.ContextCreate(params)
	if err != nil {
		log.Printf("Error: %s", err)
		return false
	}
	//log.Printf("%+v\n", ctx)

	signature := append(r.Bytes(), s.Bytes()...)
	_, ecdsaSignature, err := secp256k1.EcdsaSignatureParseCompact(ctx, signature)
	if err != nil {
		log.Printf("Error: %s", err)
		return false
	}

	publicKey := make([]byte, 1)
	publicKey[0] = 0x04
	publicKey = append(publicKey, pub.X.Bytes()...)
	publicKey = append(publicKey, pub.Y.Bytes()...)
	//log.Printf("Public Key: %s\n", hex.EncodeToString(publicKey))
	_, publicKeyStruct, err := secp256k1.EcPubkeyParse(ctx, publicKey)
	if err != nil {
		log.Printf("Error: %s", err)
		return false
	}

	ret, err := secp256k1.EcdsaVerify(ctx, ecdsaSignature, hash, publicKeyStruct)
	if err != nil {
		log.Printf("Error: %s", err)
		return false
	}
	//log.Println("ret", ret)

	secp256k1.ContextDestroy(ctx)

	return ret == 1
}
