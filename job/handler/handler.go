package handler

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job/queue"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/somc"
)

// Handler has all worker routines.
type Handler struct {
	abi            abi.ABI
	db             *reform.DB
	decryptKeyFunc data.ToPrivateKeyFunc
	ept            *ept.Service
	ethBack        EthBackend
	pscAddr        common.Address
	pwdGetter      data.PWDGetter
	somc           *somc.Conn
	queue          *queue.Queue
}

// NewHandler returns new instance of worker.
func NewHandler(db *reform.DB, somc *somc.Conn,
	ethBack EthBackend, pscAddr common.Address, payAddr string,
	pwdGetter data.PWDGetter,
	decryptKeyFunc data.ToPrivateKeyFunc) (*Handler, error) {

	abi, err := abi.JSON(strings.NewReader(contract.PrivatixServiceContractABI))
	if err != nil {
		return nil, err
	}

	eptService, err := ept.New(db, payAddr)
	if err != nil {
		return nil, err
	}

	return &Handler{
		abi:            abi,
		db:             db,
		decryptKeyFunc: decryptKeyFunc,
		ept:            eptService,
		ethBack:        ethBack,
		pscAddr:        pscAddr,
		pwdGetter:      pwdGetter,
		somc:           somc,
	}, nil
}

// SetQueue sets queue for handlers.
func (h *Handler) SetQueue(queue *queue.Queue) {
	h.queue = queue
}
