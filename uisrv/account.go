package uisrv

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if c := strings.Split(r.URL.Path, "/"); len(c) > 1 {
		id, format := c[len(c)-2], c[len(c)-1]
		if r.Method == http.MethodGet && format == "pkey" {
			s.handleExportAccount(w, r, id)
			return
		}
		if r.Method == http.MethodPut && format == "status" {
			s.handleAccountTransferBalance(w, r, id)
			return
		}
		if r.Method == http.MethodPost && format == "balances-update" {
			s.handleCreateUpdateBalancesJob(w, r, id)
			return
		}
	}

	if r.Method == http.MethodGet {
		s.handleGetAccounts(w, r)
		return
	}

	if r.Method == http.MethodPost {
		s.handleCreateAccount(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) handleExportAccount(w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handleExportAccount")
	if !util.IsUUID(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var acc data.Account
	if err := s.db.FindByPrimaryKeyTo(&acc, id); err != nil {
		if err == reform.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	privKeyJSONBytes, err := data.ToBytes(acc.PrivateKey)
	if err != nil {
		s.replyUnexpectedErr(logger, w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(privKeyJSONBytes); err != nil {
		logger.Warn(fmt.Sprintf(
			"failed to reply with the private key: %v", err))
	}
}

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{{Name: "id", Field: "id"}},
		View:   data.AccountTable,
	})
}

type accountCreatePayload struct {
	PrivateKey           string `json:"privateKey"`
	JSONKeyStoreRaw      string `json:"jsonKeyStoreRaw"`
	JSONKeyStorePassword string `json:"jsonKeyStorePassword"`
	IsDefault            bool   `json:"isDefault"`
	InUse                bool   `json:"inUse"`
	Name                 string `json:"name"`
}

func (p *accountCreatePayload) fromPrivateKeyToECDSA() (*ecdsa.PrivateKey, error) {
	pkBytes, err := data.ToBytes(p.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode private key: %v", err)
	}
	privKey, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return nil, fmt.Errorf("could not make ecdsa priv key: %v", err)
	}
	return privKey, nil
}

func (p *accountCreatePayload) fromJSONKeyStoreRawToECDSA() (*ecdsa.PrivateKey, error) {
	key, err := keystore.DecryptKey([]byte(p.JSONKeyStoreRaw), p.JSONKeyStorePassword)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt keystore: %v", err)
	}
	return key.PrivateKey, nil
}

func (p *accountCreatePayload) toECDSA() (*ecdsa.PrivateKey, error) {
	if p.PrivateKey != "" {
		return p.fromPrivateKeyToECDSA()
	} else if p.JSONKeyStoreRaw != "" {
		return p.fromJSONKeyStoreRawToECDSA()
	}

	return nil, fmt.Errorf("neither private key nor raw keystore json provided")
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Add("method", "handleCreateAccount")

	payload := &accountCreatePayload{}
	if !s.parsePayload(logger, w, r, payload) {
		return
	}
	acc := &data.Account{}
	acc.ID = util.NewUUID()

	privKey, err := payload.toECDSA()
	if err != nil {
		logger.Warn(fmt.Sprintf("could not extract priv key: %v", err))
		s.replyInvalidRequest(logger, w)
		return
	}

	acc.PrivateKey, err = s.encryptKeyFunc(privKey, s.pwdStorage.Get())
	if err != nil {
		logger.Warn(fmt.Sprintf("could not encrypt priv key: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	acc.PublicKey = data.FromBytes(crypto.FromECDSAPub(&privKey.PublicKey))

	ethAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	acc.EthAddr = data.HexFromBytes(ethAddr.Bytes())

	acc.IsDefault = payload.IsDefault
	acc.InUse = payload.InUse
	acc.Name = payload.Name

	// Set 0 balances on initial create.
	acc.PTCBalance = 0
	acc.PSCBalance = 0
	acc.EthBalance = data.B64BigInt(data.FromBytes([]byte{0}))

	if err := s.db.Insert(acc); err != nil {
		logger.Warn(fmt.Sprintf("could not insert account: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	if err := s.queue.Add(&data.Job{
		RelatedType: data.JobAccount,
		RelatedID:   acc.ID,
		Type:        data.JobAccountUpdateBalances,
		CreatedBy:   data.JobUser,
		Data:        []byte("{}"),
	}); err != nil {
		logger.Error(fmt.Sprintf("could not add %s job",
			data.JobAccountUpdateBalances))
		s.replyUnexpectedErr(logger, w)
		return
	}

	s.replyEntityCreated(logger, w, acc.ID)
}

type accountBalancePayload struct {
	Amount      uint64 `json:"amount"`
	Destination string `json:"destination"`
	GasPrice    uint64 `json:"gasPrice"`
}

func (s *Server) handleAccountTransferBalance(
	w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handleAccountTransferBalance")

	payload := &accountBalancePayload{}
	if !s.parsePayload(logger, w, r, payload) {
		return
	}
	if payload.Amount == 0 || (payload.Destination != data.ContractPSC &&
		payload.Destination != data.ContractPTC) {
		s.replyErr(logger, w, http.StatusBadRequest, &serverError{
			Message: "invalid amount or destination",
		})
		return
	}

	if !s.findTo(logger, w, &data.Account{}, id) {
		return
	}

	jobType := data.JobPreAccountAddBalanceApprove
	if payload.Destination == data.ContractPTC {
		jobType = data.JobPreAccountReturnBalance
	}

	jobData := &data.JobBalanceData{
		Amount:   payload.Amount,
		GasPrice: payload.GasPrice,
	}

	jobDataB, err := json.Marshal(jobData)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to marshal %T: %v", jobData, err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	if err = s.queue.Add(&data.Job{
		Type:        jobType,
		RelatedType: data.JobAccount,
		RelatedID:   id,
		Data:        jobDataB,
		CreatedBy:   data.JobUser,
	}); err != nil {
		logger.Error(
			fmt.Sprintf("failed to add transfer job: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
}

func (s *Server) handleCreateUpdateBalancesJob(
	w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handleCreateUpdateBalancesJob", "id", id)

	if !s.findTo(logger, w, &data.Account{}, id) {
		return
	}

	if err := s.queue.Add(&data.Job{
		Type:        data.JobAccountUpdateBalances,
		RelatedType: data.JobAccount,
		RelatedID:   id,
		Data:        []byte("{}"),
		CreatedBy:   data.JobUser,
	}); err != nil {
		s.logger.Error(
			fmt.Sprintf("failed to add update balances job: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
}
