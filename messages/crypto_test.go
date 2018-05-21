package messages_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"os"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/util"
)

var testPassword = "test-password"

func TestMain(m *testing.M) {
	// Ignore config when all tests run.
	util.ReadTestConfig(&struct{}{})

	os.Exit(m.Run())
}

func TestSealOpen(t *testing.T) {
	msg := []byte(`{"foo": "bar"}`)

	clientKey, _ := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)
	agentKey, _ := ecdsa.GenerateKey(ethcrypto.S256(), rand.Reader)

	sealed, err := messages.AgentSeal(msg,
		ethcrypto.FromECDSAPub(&clientKey.PublicKey), agentKey)
	if err != nil {
		t.Fatal("failed to encrypt: ", err)
	}

	opened, err := messages.ClientOpen(sealed,
		ethcrypto.FromECDSAPub(&agentKey.PublicKey), clientKey)
	if err != nil {
		t.Fatal("failed to decrypt: ", err)
	}

	if !bytes.Equal(opened, msg) {
		t.Fatalf("got: %x, want: %x", opened, msg)
	}
}
