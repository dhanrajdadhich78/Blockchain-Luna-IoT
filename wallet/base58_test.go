package wallet

import (
"testing"
s "wizeBlockchain/services"
"github.com/stretchr/testify/assert"
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
