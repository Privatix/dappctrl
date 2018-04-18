// +build !noethtest

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// GethNode specifies config for remote geth node interface.
type GethNode struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// Interface returns http scheme for accessing geth JSON RPC API
func (g *GethNode) Interface() string {
	return fmt.Sprint("http://", g.Host, ":", g.Port)
}

// TruffleAPI specifies config for remote TruffleAPI interface (tests environment).
type TruffleAPI struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

// Interface returns http scheme for accessing truffle API (tests environment)
func (t *TruffleAPI) Interface() string {
	return fmt.Sprint("http://", t.Host, ":", t.Port)
}

// EthereumConf specifies config for ethereum communication.
type EthereumConf struct {
	Geth       GethNode   `json:"geth"`
	TruffleAPI TruffleAPI `json:"truffle"`
}

var (
	conf *EthereumConf = nil
)

// GethEthereumConfig returns ethereum configuration.
func GethEthereumConfig() *EthereumConf {
	if conf == nil {
		loadTestConfig()
	}

	return conf
}

func loadTestConfig() {
	confPath, err := filepath.Abs(filepath.FromSlash("../../dappctrl-test.config.json"))
	if err != nil {
		panic(err)
	}

	confData, err := ioutil.ReadFile(confPath)
	if err != nil {
		panic(err)
	}

	data := make(map[string]json.RawMessage)
	err = json.Unmarshal(confData, &data)
	if err != nil {
		panic(err)
	}

	conf = &EthereumConf{}
	err = json.Unmarshal(data["EthereumLib"], conf)
	if err != nil {
		panic(err)
	}
}
