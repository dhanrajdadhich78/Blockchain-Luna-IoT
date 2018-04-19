package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func testGenerateKeyPair(t *testing.T) {
	private, err := GenerateKey(nil, rand.Reader)
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Private Key: %x\n", private.D.Bytes())
	t.Logf("Public Key: %x - %x\n", private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes())
}

func testSign(t *testing.T) {
	private, err := GenerateKey(nil, rand.Reader)
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Private Key: %x\n", private.D.Bytes())
	t.Logf("Public Key:  %x - %x\n", private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes())

	stringToSign := "Test"
	hash := sha256.Sum256([]byte(stringToSign))
	t.Logf("Hash:        %x\n", hash)

	r, s, err := Sign(rand.Reader, private, hash[:])
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Signature:   %x - %x\n", r.Bytes(), s.Bytes())
}

func TestSignAndVerify(t *testing.T) {
	private, err := GenerateKey(nil, rand.Reader)
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Private Key: %x\n", private.D.Bytes())
	t.Logf("Public Key:  %x - %x\n", private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes())

	stringToSign := "Test"
	hash := sha256.Sum256([]byte(stringToSign))
	t.Logf("Hash:        %x\n", hash)

	r, s, err := Sign(rand.Reader, private, hash[:])
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Signature:   %x - %x\n", r.Bytes(), s.Bytes())

	ret := Verify(&private.PublicKey, hash[:], r, s)
	if ret {
		t.Logf("Verify Passed")
	} else {
		t.Errorf("Verify Failed")
	}
}
