package ui

import (
	"crypto/ecdsa"
	"encoding/json"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// AccountParams is format of input to create an account.
type AccountParams struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
	InUse     bool   `json:"inUse"`
}

// AccountParamsWithHexKey is format of input to create account with given key.
type AccountParamsWithHexKey struct {
	AccountParams
	PrivateKeyHex string `json:"privateKeyHex"`
}

func (p *AccountParams) prefilledAccount() *data.Account {
	if p == nil {
		return &data.Account{}
	}
	return &data.Account{
		Name:      p.Name,
		IsDefault: p.IsDefault,
		InUse:     p.InUse,
	}
}

// ExportPrivateKey returns a private key in base64 encoding by account id.
func (h *Handler) ExportPrivateKey(
	password, account string) ([]byte, error) {
	logger := h.logger.Add("method", "ExportPrivateKey",
		"account", account)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	var acc data.Account
	err := h.db.FindByPrimaryKeyTo(&acc, account)
	if err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return nil, ErrAccountNotFound
		}
		return nil, ErrInternal
	}
	key, err := data.ToBytes(acc.PrivateKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return key, nil
}

// GetAccounts returns accounts.
func (h *Handler) GetAccounts(password string) ([]data.Account, error) {
	logger := h.logger.Add("method", "GetAccounts")

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	accounts, err := h.selectAllFrom(logger, data.AccountTable, "")
	if err != nil {
		return nil, err
	}

	result := make([]data.Account, len(accounts))

	for k, v := range accounts {
		result[k] = *v.(*data.Account)
	}

	return result, nil
}

func (h *Handler) hexPrivateKeyToECDSA(
	privateKey string) func() (*ecdsa.PrivateKey, error) {
	logger := h.logger.Add("method", "hexPrivateKeyToECDSA")

	return func() (*ecdsa.PrivateKey, error) {
		pkBytes, err := data.HexToBytes(privateKey)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrFailedToDecodePrivateKey
		}

		key, err := crypto.ToECDSA(pkBytes)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrInternal
		}
		return key, nil
	}
}

func (h *Handler) jsonPrivateKeyToECDSA(jsonBlob json.RawMessage,
	password string) func() (*ecdsa.PrivateKey, error) {
	logger := h.logger.Add("method", "jsonPrivateKeyToECDSA")

	return func() (*ecdsa.PrivateKey, error) {
		key, err := keystore.DecryptKey(
			jsonBlob, password)
		if err != nil {
			logger.Error(err.Error())
			return nil, ErrFailedToDecryptPKey
		}
		return key.PrivateKey, nil
	}
}

func (h *Handler) fillAndSaveAccount(logger log.Logger, account *data.Account,
	makeECDSAFunc func() (*ecdsa.PrivateKey, error),
	updateBalances bool) (string, error) {
	account.ID = util.NewUUID()

	pk, err := makeECDSAFunc()
	if err != nil {
		logger.Error(err.Error())
		return "", ErrInternal
	}

	account.PrivateKey, err = h.encryptKeyFunc(pk, h.pwdStorage.Get())
	if err != nil {
		logger.Error(err.Error())
		return "", ErrInternal
	}

	account.PublicKey = data.FromBytes(crypto.FromECDSAPub(&pk.PublicKey))

	ethAddr := crypto.PubkeyToAddress(pk.PublicKey)
	account.EthAddr = data.HexFromBytes(ethAddr.Bytes())

	// Set 0 balances on initial create.
	account.PTCBalance = 0
	account.PSCBalance = 0
	account.EthBalance = data.B64BigInt(data.FromBytes([]byte{0}))

	err = insert(logger, h.db.Querier, account)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}

	if updateBalances {
		err = job.AddSimple(h.queue, nil, data.JobAccountUpdateBalances,
			data.JobAccount, account.ID, data.JobUser)
		if err != nil {
			logger.Error(err.Error())
			return "", ErrInternal
		}
	}

	return account.ID, nil
}

// GenerateAccount generates new private key and creates new account.
func (h *Handler) GenerateAccount(
	password string, params *AccountParams) (*string, error) {
	logger := h.logger.Add("method", "GenerateAccount")

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	account := params.prefilledAccount()

	id, err := h.fillAndSaveAccount(
		logger, account, crypto.GenerateKey, false)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// ImportAccountFromHex imports private key from hex, creates account
// and initiates JobAccountUpdateBalances job.
func (h *Handler) ImportAccountFromHex(
	password string, params *AccountParamsWithHexKey) (*string, error) {
	logger := h.logger.Add("method", "ImportAccountFromHex")

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	makeECDSAFunc := h.hexPrivateKeyToECDSA(params.PrivateKeyHex)

	account := params.prefilledAccount()

	id, err := h.fillAndSaveAccount(logger, account, makeECDSAFunc, true)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// ImportAccountFromJSON imports private key from JSON blob with password,
// creates account and initiates JobAccountUpdateBalances job.
func (h *Handler) ImportAccountFromJSON(
	password string, params *AccountParams, jsonBlob json.RawMessage,
	jsonKeyStorePassword string) (*string, error) {
	logger := h.logger.Add("method", "ImportAccountFromJSON")

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	account := params.prefilledAccount()

	makeECDSAFunc := h.jsonPrivateKeyToECDSA(
		jsonBlob, jsonKeyStorePassword)

	id, err := h.fillAndSaveAccount(
		logger, account, makeECDSAFunc, true)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// TransferTokens initiates JobPreAccountAddBalanceApprove
// or JobPreAccountReturnBalance job depending on the direction of the transfer.
func (h *Handler) TransferTokens(
	password, account, destination string, amount, gasPrice uint64) error {
	logger := h.logger.Add("method", "TransferTokens", "destination",
		destination, "amount", amount, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	if amount == 0 {
		logger.Error(ErrTokenAmountTooSmall.Error())
		return ErrTokenAmountTooSmall
	}

	if destination != data.ContractPSC && destination != data.ContractPTC {
		logger.Error(ErrBadDestination.Error())
		return ErrBadDestination
	}

	err := h.findByPrimaryKey(
		logger, ErrAccountNotFound, &data.Account{}, account)
	if err != nil {
		return err
	}

	jobType := data.JobPreAccountAddBalanceApprove
	if destination == data.ContractPTC {
		jobType = data.JobPreAccountReturnBalance
	}

	jobData := &data.JobBalanceData{
		Amount:   amount,
		GasPrice: gasPrice,
	}

	err = job.AddWithData(h.queue, nil, jobType,
		data.JobAccount, account, data.JobUser, jobData)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// UpdateBalance initiates JobAccountUpdateBalances job.
func (h *Handler) UpdateBalance(password, account string) error {
	logger := h.logger.Add("method", "UpdateBalance",
		"account", account)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	err := h.findByPrimaryKey(
		logger, ErrAccountNotFound, &data.Account{}, account)
	if err != nil {
		return err
	}

	err = job.AddSimple(h.queue, nil, data.JobAccountUpdateBalances,
		data.JobAccount, account, data.JobUser)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// UpdateAccount updates an account.
func (h *Handler) UpdateAccount(password, account, name string,
	isDefault, inUse bool) error {
	logger := h.logger.Add("method", "UpdateAccount",
		"account", account)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	acc := data.Account{}
	err := h.findByPrimaryKey(
		logger, ErrAccountNotFound, &acc, account)
	if err != nil {
		return err
	}

	if name != "" {
		acc.Name = name
	}

	acc.IsDefault = isDefault
	acc.InUse = inUse

	return update(logger, h.db.Querier, &acc)
}
