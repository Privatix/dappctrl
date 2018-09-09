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
)

// ExportPrivateKeyResult is an ExportPrivateKey result.
type ExportPrivateKeyResult struct {
	PrivateKey []byte `json:"privateKey"`
}

// GetAccountsResult is an GetAccounts result.
type GetAccountsResult struct {
	Accounts []data.Account
}

// CreateAccountResult is an CreateAccount result.
type CreateAccountResult struct {
	Account string `json:"account"`
}

// ExportPrivateKey returns a private key in base64 encoding by account id.
func (h *Handler) ExportPrivateKey(
	password, account string) (*ExportPrivateKeyResult, error) {
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

	return &ExportPrivateKeyResult{key}, nil
}

// GetAccounts returns accounts.
func (h *Handler) GetAccounts(password string) (*GetAccountsResult, error) {
	logger := h.logger.Add("method", "GetAccounts")

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	accounts, err := h.selectAllFrom(logger, data.AccountTable, "")
	if err != nil {
		return nil, err
	}

	result := &GetAccountsResult{}

	for _, v := range accounts {
		result.Accounts = append(result.Accounts, *v.(*data.Account))
	}

	return result, nil
}

func (h *Handler) fromPrivateKeyToECDSA(
	privateKey string) (*ecdsa.PrivateKey, error) {
	logger := h.logger.Add("method", "fromPrivateKeyToECDSA")

	pkBytes, err := data.ToBytes(privateKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrDecodePrivateKey
	}

	key, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return key, nil
}

func (h *Handler) fromJSONKeyStoreRawToECDSA(jsonKeyStoreRaw json.RawMessage,
	jsonKeyStorePassword string) (*ecdsa.PrivateKey, error) {
	logger := h.logger.Add("method", "fromJSONKeyStoreRawToECDSA")

	key, err := keystore.DecryptKey(
		jsonKeyStoreRaw, jsonKeyStorePassword)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrDecryptKeystore
	}
	return key.PrivateKey, nil
}

func (h *Handler) toECDSA(privateKey, jsonKeyStorePassword string,
	jsonKeyStoreRaw json.RawMessage) (*ecdsa.PrivateKey, error) {
	logger := h.logger.Add("method", "toECDSA")

	if privateKey != "" {
		return h.fromPrivateKeyToECDSA(privateKey)
	} else if len(jsonKeyStoreRaw) != 0 {
		return h.fromJSONKeyStoreRawToECDSA(
			jsonKeyStoreRaw, jsonKeyStorePassword)
	}

	logger.Error(ErrPrivateKeyNotFound.Error())

	return nil, ErrPrivateKeyNotFound
}

// CreateAccount creates new account and initiates JobAccountUpdateBalances job.
func (h *Handler) CreateAccount(password, privateKey string,
	jsonKeyStoreRaw json.RawMessage, jsonKeyStorePassword, name string,
	isDefault, inUse bool) (*CreateAccountResult, error) {
	logger := h.logger.Add("method", "CreateAccount",
		"name", name, "isDefault", isDefault, "inUse", inUse)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	account := &data.Account{}
	account.ID = util.NewUUID()

	pk, err := h.toECDSA(privateKey, jsonKeyStorePassword, jsonKeyStoreRaw)
	if err != nil {
		return nil, err
	}

	account.PrivateKey, err = h.encryptKeyFunc(pk, h.pwdStorage.Get())
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	account.PublicKey = data.FromBytes(crypto.FromECDSAPub(&pk.PublicKey))

	ethAddr := crypto.PubkeyToAddress(pk.PublicKey)
	account.EthAddr = data.HexFromBytes(ethAddr.Bytes())

	account.IsDefault = isDefault
	account.InUse = inUse
	account.Name = name

	// Set 0 balances on initial create.
	account.PTCBalance = 0
	account.PSCBalance = 0
	account.EthBalance = data.B64BigInt(data.FromBytes([]byte{0}))

	err = h.db.Insert(account)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	err = job.AddSimple(h.queue, data.JobAccountUpdateBalances,
		data.JobAccount, account.ID, data.JobUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return &CreateAccountResult{account.ID}, nil
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
		logger.Error(ErrSmallTokensAmount.Error())
		return ErrSmallTokensAmount
	}

	if destination != data.ContractPSC && destination != data.ContractPTC {
		logger.Error(ErrDestination.Error())
		return ErrDestination
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

	err = job.AddWithData(h.queue, jobType,
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

	err = job.AddSimple(h.queue, data.JobAccountUpdateBalances,
		data.JobAccount, account, data.JobUser)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}
