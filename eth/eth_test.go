// +build !noethtest

package eth

import (
	"os"
	"testing"

	"github.com/privatix/dappctrl/eth/truffle"
	"github.com/privatix/dappctrl/util"
)

var (
	testGethURL    string
	testTruffleAPI truffle.API
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
	os.Exit(m.Run())
}

func getClient() *EthereumClient {
	return NewEthereumClient(testGethURL)
}
