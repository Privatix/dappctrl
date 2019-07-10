package data_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
)

func TestPrivateKeyEncrypttionAndDecryption(t *testing.T) {
	auth := "test-passphrase"

	pkey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal("failed to generate ecdsa.PrivateKey: ", err)
	}

	encrypted, err := data.EncryptedKey(pkey, auth)
	if err != nil {
		t.Fatal("failed to encrypt: ", err)
	}

	pkeyDecr, err := data.ToPrivateKey(encrypted, auth)
	if err != nil {
		t.Fatal("failed to decrypt: ", err)
	}

	if !bytes.Equal(crypto.FromECDSA(pkey), crypto.FromECDSA(pkeyDecr)) {
		t.Fatal("initial and decrypted keys do not match")
	}
}
