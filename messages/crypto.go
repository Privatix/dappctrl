package messages

import (
	"crypto/ecdsa"
	"crypto/rand"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

const sigLen = 64

// AgentSeal encrypts message using client's public key and packs with
// agent signature.
func AgentSeal(msg, clientPub []byte, agentKey *ecdsa.PrivateKey) ([]byte, error) {
	pubKey, err := ethcrypto.UnmarshalPubkey(clientPub)
	if err != nil {
		return nil, err
	}

	pub := ecies.ImportECDSAPublic(pubKey)
	msgEncrypted, err := ecies.Encrypt(rand.Reader, pub, msg, nil, nil)
	if err != nil {
		return nil, err
	}

	return PackWithSignature(msgEncrypted, agentKey)
}

// ClientOpen decrypts message using client's key and verifies using agent's key.
func ClientOpen(c, agentPub []byte, clientPrv *ecdsa.PrivateKey) ([]byte, error) {
	sealed, sig := UnpackSignature(c)
	hash := ethcrypto.Keccak256(sealed)

	if !VerifySignature(agentPub, hash, sig) {
		return nil, ErrWrongSignature
	}

	prv := ecies.ImportECDSA(clientPrv)

	opened, err := prv.Decrypt(sealed, nil, nil)
	if err != nil {
		return nil, err
	}

	return opened, nil
}

// PackWithSignature packs message with signature.
func PackWithSignature(msg []byte, key *ecdsa.PrivateKey) ([]byte, error) {
	sig, err := signature(key, msg)
	if err != nil {
		return nil, err
	}

	return packSignature(msg, sig), nil
}

// UnpackSignature unpacks msg from signature.
func UnpackSignature(c []byte) (msg []byte, sig []byte) {
	msg = c[:len(c)-sigLen]
	sig = c[len(c)-sigLen:]
	return
}

// VerifySignature returns true if signature is correct.
func VerifySignature(pubk, msg, sig []byte) bool {
	return ethcrypto.VerifySignature(pubk, msg, sig)
}

// signature computes and returns signature.
func signature(key *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	hash := ethcrypto.Keccak256(msg)
	sig, err := ethcrypto.Sign(hash, key)
	if err != nil {
		return nil, err
	}
	sig = sig[:len(sig)-1]
	return sig, nil
}

func packSignature(msg, sig []byte) []byte {
	ret := make([]byte, len(msg)+len(sig))
	copy(ret, msg)
	copy(ret[len(msg):], sig)

	return ret
}
