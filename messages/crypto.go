package messages

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

const sigLen = 64

// AgentSeal encrypts message using client's public key and packs with
// agent signature.
func AgentSeal(msg, clientPub []byte, agentKey *ecdsa.PrivateKey) ([]byte, error) {
	pub := ecies.ImportECDSAPublic(ethcrypto.ToECDSAPub(clientPub))
	msgEncrypted, err := ecies.Encrypt(rand.Reader, pub, msg, nil, nil)
	if err != nil {
		return nil, err
	}

	sig, err := Signature(agentKey, msgEncrypted)
	if err != nil {
		return nil, err
	}

	return PackSignature(msgEncrypted, sig), nil
}

// PackSignature packs msg with signature.
func PackSignature(msg, sig []byte) []byte {
	ret := make([]byte, len(msg)+len(sig))
	copy(ret, msg)
	copy(ret[len(msg):], sig)

	return ret
}

// Signature computes and returns signature.
func Signature(key *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	hash := ethcrypto.Keccak256(msg)
	sig, err := ethcrypto.Sign(hash, key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}
	sig = sig[:len(sig)-1]
	return sig, nil
}

// ClientOpen decrypts message using client's key and verifies using agent's key.
func ClientOpen(c, agentPub []byte, clientPrv *ecdsa.PrivateKey) ([]byte, error) {
	sealed := c[:len(c)-sigLen]
	hash := ethcrypto.Keccak256(sealed)
	sig := c[len(c)-sigLen:]

	if !ethcrypto.VerifySignature(agentPub, hash, sig) {
		return nil, fmt.Errorf("wrong signature")
	}

	prv := ecies.ImportECDSA(clientPrv)

	opened, err := prv.Decrypt(sealed, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return opened, nil
}
