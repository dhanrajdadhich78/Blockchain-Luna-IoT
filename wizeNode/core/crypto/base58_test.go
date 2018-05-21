package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase58(t *testing.T) {
	for i := 0; i < 1000; i++ {
		_, public := NewKeyPair()
		pubKeyHash := HashPubKey(public)
		//t.Logf("pubKeyHash: %x", pubKeyHash)

		versionedPayload := append([]byte{Version}, pubKeyHash...)
		checksum := Checksum(versionedPayload)

		fullPayload := append(versionedPayload, checksum...)
		address := Base58Encode(fullPayload)
		//t.Logf("address: %s", address)

		assert.Equal(
			t,
			ValidateAddress(string(address[:])),
			true,
			"Address: %s is invalid", address,
		)
	}
}

func TestBase58IfPubKeyHashStartsWith00(t *testing.T) {
	pubKeyHash, _ := hex.DecodeString("0034a86a344396b04b31209bbeb6e91596b59e0c")
	address := GetAddressFromPubKeyHash(pubKeyHash)
	//t.Logf("address: %s", address)

	assert.Equal(
		t,
		ValidateAddress(string(address[:])),
		true,
		"Address: %s is invalid", address,
	)
}
