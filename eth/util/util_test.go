package util

import (
	"bytes"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	auth := bind.NewKeyedTransactor(key)

	tx := types.NewTransaction(0, common.HexToAddress("0x11"), big.NewInt(1),
		0, big.NewInt(1), []byte{})

	signer := &types.HomesteadSigner{}

	signedTx, err := auth.Signer(signer, auth.From, tx)
	if err != nil {
		t.Fatal(err)
	}

	pubk, err := RecoverPubKey(signer, signedTx)
	if err != nil {
		t.Fatal("failed to recover: ", err)
	}

	expectedB := crypto.FromECDSAPub(&key.PublicKey)
	actualB := crypto.FromECDSAPub(pubk)
	if !bytes.Equal(expectedB, actualB) {
		t.Fatal("wrong pub key recovered")
	}
}
