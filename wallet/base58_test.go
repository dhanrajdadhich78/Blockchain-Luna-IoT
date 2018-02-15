package wallet

import (
	"github.com/stretchr/testify/assert"
	"testing"
	s "wizeBlock/services"
)

func TestBase58(t *testing.T) {
	for i := 0; i < 100; i++ {
		_, public := NewKeyPair()
		pubKeyHash := HashPubKey(public)

		versionedPayload := append([]byte{Version}, pubKeyHash...)
		checksum := Checksum(versionedPayload)

		fullPayload := append(versionedPayload, checksum...)
		address := s.Base58Encode(fullPayload)

		assert.Equal(
			t,
			ValidateAddress(string(address[:])),
			true,
			"Address: %s is invalid", address,
		)
	}
}
