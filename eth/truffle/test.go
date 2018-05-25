// +build !notest

package truffle

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type truffleAccount struct {
	Account    string `json:"account"`
	PrivateKey string `json:"privateKey"`
}

// TestAccount is a parsed account created by truffle.
type TestAccount struct {
	Address    common.Address
	PrivateKey *ecdsa.PrivateKey
}

// GetTestAccounts returns all available accounts in truffle.
func (api *API) GetTestAccounts() ([]TestAccount, error) {
	response := []truffleAccount{}
	api.fetchFromTruffle("/getKeys", &response)
	accounts := []TestAccount{}
	for _, acc := range response {
		keyBytes, err := hex.DecodeString(acc.PrivateKey)
		if err != nil {
			return nil, err
		}
		key, err := crypto.ToECDSA(keyBytes)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, TestAccount{
			Address:    common.HexToAddress(acc.Account),
			PrivateKey: key,
		})
	}
	return accounts, nil
}
