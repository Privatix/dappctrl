// +build !noethtest

package eth

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/eth/truffle"
	"github.com/privatix/dappctrl/util"
)

var (
	testGethURL    string
	testTruffleAPI truffle.API
	testEthClient  *ethclient.Client
)

// TestMain reads config and run tests.
func TestMain(m *testing.M) {
	var conf struct {
		Eth struct {
			GethURL       string
			TruffleAPIURL string
		}
	}
	util.ReadTestConfig(&conf)
	testGethURL = conf.Eth.GethURL
	testTruffleAPI = truffle.API(conf.Eth.TruffleAPIURL)
	client, err := ethclient.Dial(testGethURL)
	if err != nil {
		panic(err)
	}
	testEthClient = client
	os.Exit(m.Run())
}

func getClient() *EthereumClient {
	return NewEthereumClient(testGethURL)
}

func getPTC(t *testing.T) *contract.PrivatixTokenContract {
	ptcAddr := testTruffleAPI.FetchPTCAddress()
	ptc, err := contract.NewPrivatixTokenContract(common.HexToAddress(ptcAddr), testEthClient)
	if err != nil {
		t.Fatal("failed to create ptc instance: ", err)
	}
	return ptc
}

func getTransactorForAccount(t *testing.T, acc *truffle.TestAccount) *bind.TransactOpts {
	return bind.NewKeyedTransactor(acc.PrivateKey)
}
