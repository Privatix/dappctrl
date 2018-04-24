package keystore

import (
	"encoding/hex"
	"errors"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"math/big"
	"os"
)

// AccountsManagerConf stores configuration, related to accounts manager.
type AccountsManagerConf struct {
	// Specifies path to the dir with encrypted private keystore of accounts.
	KeystorePath string `json:"keystorePath"`
}

// AccountsManager provides safe access to the private key(s) of ethereum account.
//
// By default, current implementation enforces only one account at a time,
// due to the current bossiness model, but there is a possibility of wider count of private keystore support.
// (internally used keystore supports multiple accounts).
type AccountsManager struct {
	keystore *keystore.KeyStore
}

// NewAccountsManager returns account manager, configured with "conf".
// It doesn't perform any checks of "conf" correctness.
// In case of invalid "conf" - other methods might report corresponding error in runtime.
func NewAccountsManager(conf *AccountsManagerConf) *AccountsManager {
	return &AccountsManager{
		keystore: keystore.NewKeyStore(
			conf.KeystorePath,
			
			// More details about used password encryption algorithm:
			// https://github.com/Tarsnap/scrypt
			// Current realisation uses "StandardScryptN" and "StandardScryptP"
			// used in ethereum, by default.
			keystore.StandardScryptN,
			keystore.StandardScryptP),
	}
}

// SetPrivateKeyFromHex creates account (private and public keystore pair) from private key in hex format.
// Format example: "adfd7c50906423a1bc64f9927f5c05f2f8cdd815ffed78534ff2a1f15cbcae19"
// This is the most used key format in ethereum, many wallets uses it as default format.
func (m *AccountsManager) SetPrivateKeyFromHex(pKeyHex, passPhrase string) (accounts.Account, error) {
	if len(m.keystore.Accounts()) != 0 {
		return accounts.Account{}, errors.New(
			"only one private key at a time is supported. " +
				"please, remove already present one before initialising")
	}

	pKeyBytes, err := hex.DecodeString(pKeyHex)
	if err != nil {
		return accounts.Account{}, err
	}

	pKeyECDSA, err := crypto.ToECDSA(pKeyBytes)
	if err != nil {
		return accounts.Account{}, err
	}

	return m.keystore.ImportECDSA(pKeyECDSA, passPhrase)
}

// SetPrivateKeyFromJSON creates account (private and public keystore pair) from enc-json format.
// Used by geth by default. This format stores key and related constants in JSON file.
// For the format details, please, see internally stored JSON file in keystore dir.
func (m *AccountsManager) SetPrivateKeyFromJSON(filePath string, passPhrase string) (accounts.Account, error) {
	if len(m.keystore.Accounts()) != 0 {
		return accounts.Account{}, errors.New(
			"only one private key at a time is supported. " +
				"please, remove already present one before initialising")
	}

	jsonKey, err := ioutil.ReadFile(filePath)
	if err != nil {
		return accounts.Account{}, err
	}

	return m.keystore.Import(jsonKey, passPhrase, passPhrase)
}

// UpdatePassword encrypts account with new password.
// Returns error in case if previous password is invalid or there are any other internal error.
func (m *AccountsManager) UpdatePassword(previousPassPhrase, newPassPhrase string) error {
	err := m.checkAccountPresence()
	if err != nil {
		return err
	}

	account := m.keystore.Accounts()[0]
	return m.keystore.Update(account, previousPassPhrase, newPassPhrase)
}

// Transactor returns transaction signer object, that might be used for signing contract calls.
func (m *AccountsManager) Transactor(passPhrase string) (*bind.TransactOpts, error) {
	err := m.checkAccountPresence()
	if err != nil {
		return nil, err
	}

	keyFilePath := m.keystore.Accounts()[0].URL.Path
	file, err := os.Open(keyFilePath)
	if err != nil {
		return nil, err
	}

	return bind.NewTransactor(file, passPhrase)
}

// SignTransaction signs transaction "tx" with private key.
// Internally it uses default method of ethereum's keystore - SignTransactionWithPassword.
// This wrapper only ensures account presence before signing attempt.
func (m *AccountsManager) SignTransaction(tx *types.Transaction, chainID *big.Int, passPhrase string) (*types.Transaction, error) {
	err := m.checkAccountPresence()
	if err != nil {
		return nil, err
	}

	account := m.keystore.Accounts()[0]
	return m.keystore.SignTxWithPassphrase(account, passPhrase, tx, chainID)
}

// SignHash signs "hash" with private key.
// Internally it uses default method of ethereum's keystore - SignHashWithPassword.
// This wrapper only ensures account presence before signing attempt.
func (m *AccountsManager) SignHash(hash []byte, passPhrase string) ([]byte, error) {
	if len(m.keystore.Accounts()) == 0 {
		return nil, errors.New("no ethereum account is present yet")
	}

	account := m.keystore.Accounts()[0]
	return m.keystore.SignHashWithPassphrase(account, passPhrase, hash)
}

func (m *AccountsManager) checkAccountPresence() error {
	if len(m.keystore.Accounts()) == 0 {
		return errors.New("no ethereum account is present yet")
	}

	if len(m.keystore.Accounts()) > 1 {
		return errors.New("only one account is supported at a time")
	}

	return nil
}
