package data

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/ethereum/go-ethereum/crypto"
)

// EncryptedKeyFunc is a func that returns encrypted keystore.Key in base64.
type EncryptedKeyFunc func(*ecdsa.PrivateKey, string) (string, error)

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

// ToPrivateKeyFunc is a func that returns decrypted *ecdsa.PrivateKey from base64 of encrypted keystore.Key.
type ToPrivateKeyFunc func(string, string) (*ecdsa.PrivateKey, error)

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
