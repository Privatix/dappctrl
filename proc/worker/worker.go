package worker

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/somc"
)

// GasConf amounts of gas limit to use for contracts calls.
type GasConf struct {
	PTC struct {
		Approve uint64
	}
	PSC struct {
		AddBalanceERC20                uint64
		RegisterServiceOffering        uint64
		CreateChannel                  uint64
		CooperativeClose               uint64
		ReturnBalanceERC20             uint64
		SetNetworkFee                  uint64
		UncooperativeClose             uint64
		Settle                         uint64
		TopUp                          uint64
		GetChannelInfo                 uint64
		PublishServiceOfferingEndpoint uint64
		GetKey                         uint64
		BalanceOf                      uint64
		PopupServiceOffering           uint64
		RemoveServiceOffering          uint64
	}
}

// Worker has all worker routines.
type Worker struct {
	abi            abi.ABI
	db             *reform.DB
	decryptKeyFunc data.ToPrivateKeyFunc
	ept            *ept.Service
	ethBack        EthBackend
	gasConf        *GasConf
	pscAddr        common.Address
	pwdGetter      data.PWDGetter
	somc           *somc.Conn
	queue          *job.Queue
}

// NewWorker returns new instance of worker.
func NewWorker(db *reform.DB, somc *somc.Conn,
	ethBack EthBackend, gasConc *GasConf,
	pscAddr common.Address,
	payAddr string, pwdGetter data.PWDGetter,
	decryptKeyFunc data.ToPrivateKeyFunc) (*Worker, error) {

	abi, err := abi.JSON(strings.NewReader(contract.PrivatixServiceContractABI))
	if err != nil {
		return nil, err
	}

	eptService, err := ept.New(db, payAddr)
	if err != nil {
		return nil, err
	}

	return &Worker{
		abi:            abi,
		db:             db,
		decryptKeyFunc: decryptKeyFunc,
		gasConf:        gasConc,
		ept:            eptService,
		ethBack:        ethBack,
		pscAddr:        pscAddr,
		pwdGetter:      pwdGetter,
		somc:           somc,
	}, nil
}

// SetQueue sets queue for handlers.
func (h *Worker) SetQueue(queue *job.Queue) {
	h.queue = queue
}
