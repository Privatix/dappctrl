package uisrv

import (
	"encoding/hex"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"

	"github.com/privatix/dappctrl/util"
)

func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
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

func (s *Server) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: []queryParam{{Name: "id", Field: "id"}},
		View:   data.AccountTable,
	})
}

type accountCreatePayload struct {
	PrivateKey string `json:"privateKey"`
	IsDefault  bool   `json:"isDefault"`
	InUse      bool   `json:"inUse"`
	Name       string `json:"name"`
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	payload := &accountCreatePayload{}
	s.parsePayload(w, r, payload)
	acc := &data.Account{}
	acc.ID = util.NewUUID()
	pkBytes, err := data.ToBytes(payload.PrivateKey)
	if err != nil {
		s.logger.Warn("could not decode private key: %v", err)
		s.replyInvalidPayload(w)
		return
	}
	privKey, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		s.logger.Warn("could not make ecdsa priv key: %v", err)
		s.replyInvalidPayload(w)
		return
	}
	acc.PrivateKey, err = data.EncryptedKey(privKey, s.pwdStorage.Get())
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
