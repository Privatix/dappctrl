package worker

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/somc"
	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
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
	abi               abi.ABI
	logger            log.Logger
	db                *reform.DB
	decryptKeyFunc    data.ToPrivateKeyFunc
	ept               *ept.Service
	ethBack           eth.Backend
	gasConf           *GasConf
	pscAddr           common.Address
	pwdGetter         data.PWDGetter
	queue             job.Queue
	processor         *proc.Processor
	ethConfig         *eth.Config
	countryConfig     *country.Config
	torHostName       data.Base64String
	somcClientBuilder somc.ClientBuilderInterface
}

// NewWorker returns new instance of worker.
func NewWorker(logger log.Logger, db *reform.DB, ethBack eth.Backend,
	gasConc *GasConf, pscAddr common.Address, payAddr string,
	pwdGetter data.PWDGetter, countryConf *country.Config,
	decryptKeyFunc data.ToPrivateKeyFunc, eptConf *ept.Config,
	torHostname string, somcClientBuilder somc.ClientBuilderInterface) (*Worker, error) {

	l := logger.Add("type", "proc/worker.Worker")

	abi, err := abi.JSON(
		strings.NewReader(contract.PrivatixServiceContractABI))
	if err != nil {
		return nil, err
	}

	eptService, err := ept.New(db, logger, payAddr, eptConf.Timeout)
	if err != nil {
		return nil, err
	}

	return &Worker{
		abi:               abi,
		logger:            l,
		db:                db,
		decryptKeyFunc:    decryptKeyFunc,
		gasConf:           gasConc,
		ept:               eptService,
		ethBack:           ethBack,
		pscAddr:           pscAddr,
		pwdGetter:         pwdGetter,
		countryConfig:     countryConf,
		torHostName:       data.FromBytes([]byte(torHostname)),
		somcClientBuilder: somcClientBuilder,
	}, nil
}

// SetQueue sets a queue for handlers.
func (w *Worker) SetQueue(queue job.Queue) {
	w.queue = queue
}

// SetProcessor sets a processor.
func (w *Worker) SetProcessor(processor *proc.Processor) {
	w.processor = processor
}
