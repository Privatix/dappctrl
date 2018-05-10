package uisrv

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/reform.v1"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"

	"github.com/privatix/dappctrl/util"
)

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if c := strings.Split(r.URL.Path, "/"); len(c) > 1 {
			id, format := c[len(c)-2], c[len(c)-1]
			if format == "pkey" {
				s.handleExportAccount(w, r, id)
				return
			}
		}

		s.handleGetAccounts(w, r)
		return
	}
	if r.Method == http.MethodPost {
		s.handleCreateAccount(w, r)
		return
	}
	id := idFromStatusPath(accountsPath, r.URL.Path)
	if id != "" && r.Method == http.MethodPut {
		s.handleUpdateAccountBalance(w, r, id)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (s *Server) handleExportAccount(w http.ResponseWriter, r *http.Request, id string) {
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

	privKeyJsonBytes, err := data.ToBytes(acc.PrivateKey)
	if err != nil {
		s.replyUnexpectedErr(w)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(privKeyJsonBytes); err != nil {
		s.logger.Warn("failed to reply with the private key: %v", err)
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
	JsonKeyStoreRaw      string `json:"jsonKeyStoreRaw"`
	JsonKeyStorePassword string `json:"jsonKeyStorePassword"`
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
	key, err := keystore.DecryptKey([]byte(p.JsonKeyStoreRaw), p.JsonKeyStorePassword)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt keystore: %v", err)
	}
	return key.PrivateKey, nil
}

func (p *accountCreatePayload) toECDSA() (*ecdsa.PrivateKey, error) {
	if p.PrivateKey != "" {
		return p.fromPrivateKeyToECDSA()
	} else if p.JsonKeyStoreRaw != "" {
		return p.fromJSONKeyStoreRawToECDSA()
	}

	return nil, fmt.Errorf("neither private key nor raw keystore json provided")
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	payload := &accountCreatePayload{}
	s.parsePayload(w, r, payload)
	acc := &data.Account{}
	acc.ID = util.NewUUID()

	privKey, err := payload.toECDSA()
	if err != nil {
		s.logger.Warn("could not extract priv key: %v", err)
		s.replyInvalidPayload(w)
		return
	}

	acc.PrivateKey, err = s.encryptKeyFunc(privKey, s.pwdStorage.Get())
	if err != nil {
		s.logger.Warn("could not encrypt priv key: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	acc.PublicKey = data.FromBytes(crypto.FromECDSAPub(&privKey.PublicKey))

	ethAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	acc.EthAddr = data.FromBytes(ethAddr.Bytes())

	acc.IsDefault = payload.IsDefault
	acc.InUse = payload.InUse
	acc.Name = payload.Name

	ethAddrHex := hex.EncodeToString(ethAddr.Bytes())

	gResponse, err := s.ethClient.GetBalance("0x"+ethAddrHex, eth.BlockLatest)
	if err != nil {
		s.logger.Warn("could not get eth balance")
		s.replyUnexpectedErr(w)
		return
	}

	amount, err := eth.NewUint192(gResponse.Result)
	if err != nil {
		s.logger.Warn("could not convert geth response to uint192: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	acc.EthBalance = data.B64BigInt(data.FromBytes(amount.ToBigInt().Bytes()))

	pscBalance, err := s.psc.BalanceOf(&bind.CallOpts{}, ethAddr)
	if err != nil {
		s.logger.Warn("could not get psc balance: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	acc.PSCBalance = pscBalance.Uint64()

	ptcBalance, err := s.ptc.BalanceOf(&bind.CallOpts{}, ethAddr)
	if err != nil {
		s.logger.Warn("could not get ptc balance: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	acc.PTCBalance = ptcBalance.Uint64()

	if err := s.db.Insert(acc); err != nil {
		s.logger.Warn("could not insert account: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	s.replyEntityCreated(w, acc.ID)
}

// Actions on account's balance.
const (
	accountTransfer = "transfer"
	accountDelete   = "delete"
)

type accountBalancePayload struct {
	Action      string `json:"action"`
	Amount      uint64 `json:"amount"`
	Destination string `json:"destination"`
}

func (s *Server) handleUpdateAccountBalance(w http.ResponseWriter, r *http.Request, id string) {
	// TODO: validate request params and create balance job.
	w.WriteHeader(http.StatusBadRequest)
}
