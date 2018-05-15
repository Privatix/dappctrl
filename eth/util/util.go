package util

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// RecoverPubKey recover public key from signature values.
func RecoverPubKey(hash common.Hash, S, R, V *big.Int) (*ecdsa.PublicKey, error) {
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = byte(V.Uint64() - 27)
	// recover the public key from the signature
	pub, err := crypto.Ecrecover(hash[:], sig)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSAPub(pub), nil
}
