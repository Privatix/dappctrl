package util

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// RecoverPubKey recover public key from signature values.
func RecoverPubKey(signer types.Signer, tx *types.Transaction) (*ecdsa.PublicKey, error) {
	V, R, S := tx.RawSignatureValues()
	hash := signer.Hash(tx)
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
