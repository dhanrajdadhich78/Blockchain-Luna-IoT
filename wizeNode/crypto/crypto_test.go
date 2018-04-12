package crypto

import (
	"crypto/rand"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	private, err := GenerateKey(nil, rand.Reader)
	if err != nil {
		t.Errorf("Error: %+v\n", err)
	}

	t.Logf("Private Key: %+v\n", private)
}
