package worker

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/somc"
	reform "gopkg.in/reform.v1"
)

// Worker has all worker routines.
type Worker struct {
	db             *reform.DB
	somc           *somc.Conn
	queue          *job.Queue
	ethBack        EthBackend
	pscAddr        common.Address
	pwdGetter      data.PWDGetter
	decryptKeyFunc data.ToPrivateKeyFunc
	abi            abi.ABI
}

// NewWorker returns new instance of worker.
func NewWorker(db *reform.DB, somc *somc.Conn, queue *job.Queue,
	ethBack EthBackend, pscAddr common.Address, pwdGetter data.PWDGetter,
	decryptKeyFunc data.ToPrivateKeyFunc) (*Worker, error) {
	abi, err := abi.JSON(strings.NewReader(contract.PrivatixServiceContractABI))
	if err != nil {
		return nil, err
	}
	return &Worker{db, somc, queue, ethBack, pscAddr, pwdGetter, decryptKeyFunc, abi}, nil
}
