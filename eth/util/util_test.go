package util

import (
	"bytes"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	apputil "github.com/privatix/dappctrl/util"
)

func TestMain(m *testing.M) {
	// Ignore config flags when run all packages tests.
	apputil.ReadTestConfig(&struct{}{})
	os.Exit(m.Run())
}

func TestRecoverPublicKey(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256([]byte("test-data"))

	signature, err := crypto.Sign(hash, key)
	if err != nil {
		t.Fatal(err)
	}

	signer := types.HomesteadSigner{}
	r, s, v, err := signer.SignatureValues(nil, signature)
	if err != nil {
		t.Fatal(err)
	}

	pubk, err := RecoverPubKey(common.BytesToHash(hash), s, r, v)
	if err != nil {
		t.Fatal("failed to recover: ", err)
	}

	expectedB := crypto.FromECDSAPub(&key.PublicKey)
	actualB := crypto.FromECDSAPub(pubk)
	if !bytes.Equal(expectedB, actualB) {
		t.Fatal("wrong pub key recovered")
	}
}
