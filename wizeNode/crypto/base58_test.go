package crypto

import (
	"testing"
	b "wizeBlock/wizeNode/blockchain"

	"github.com/stretchr/testify/assert"
)

func TestBase58(t *testing.T) {
	for i := 0; i < 100; i++ {
		_, public := b.NewKeyPair()
		pubKeyHash := b.HashPubKey(public)

		versionedPayload := append([]byte{Version}, pubKeyHash...)
		checksum := b.Checksum(versionedPayload)

		fullPayload := append(versionedPayload, checksum...)
		address := Base58Encode(fullPayload)

		assert.Equal(
			t,
			ValidateAddress(string(address[:])),
			true,
			"Address: %s is invalid", address,
		)
	}
}
