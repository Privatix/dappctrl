package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

const sigLen = 64

// AgentSeal encrypts message using client's public key.
// Encrypts func is go-ethereum's ecies.Encrypt.
// Signature is generated using go-ethereum's crypto.Sign func and appended
// to the resulting message.
func AgentSeal(msg, clientPub []byte, agentPrv *ecdsa.PrivateKey) ([]byte, error) {
	pub := ecies.ImportECDSAPublic(ethcrypto.ToECDSAPub(clientPub))
	msgEncrypted, err := ecies.Encrypt(rand.Reader, pub, msg, nil, nil)
	if err != nil {
		return nil, err
	}

	hash := ethcrypto.Keccak256(msgEncrypted)
	sig, err := ethcrypto.Sign(hash, agentPrv)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %v", err)
	}
	sig = sig[:len(sig)-1]

	ret := make([]byte, len(msgEncrypted)+len(sig))
	copy(ret, msgEncrypted)
	copy(ret[len(msgEncrypted):], sig)

	return ret, nil
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
