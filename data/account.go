package data

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/ethereum/go-ethereum/crypto"
)

// EncryptedKey returns encrypted keystore.Key in base64.
func EncryptedKey(pkey *ecdsa.PrivateKey, auth string) (string, error) {
	key := keystore.NewKeyForDirectICAP(rand.Reader)
	key.Address = crypto.PubkeyToAddress(pkey.PublicKey)
	key.PrivateKey = pkey
	encryptedBytes, err := keystore.EncryptKey(key, auth,
		keystore.StandardScryptN,
		keystore.StandardScryptP)
	if err != nil {
		return "", err
	}
	return FromBytes(encryptedBytes), nil
}

// ToPrivateKey returns decrypted *ecdsa.PrivateKey from base64 of encrypted keystore.Key.
func ToPrivateKey(keyB64, auth string) (*ecdsa.PrivateKey, error) {
	keyjson, err := ToBytes(keyB64)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		return nil, err
	}
	return key.PrivateKey, nil
}

// Sign signs a data.
func (a *Account) Sign(data []byte, auth string) ([]byte, error) {
	prvKey, err := ToPrivateKey(a.PrivateKey, auth)
	if err != nil {
		return nil, err
	}
	return crypto.Sign(data, prvKey)
}
