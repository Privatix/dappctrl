// +build !noethtest

package lib

import (
	"pxctrl/eth/lib/tests"
	"testing"
)

func TestGasPriceFetching(t *testing.T) {
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)
	response, err := client.GetGasPrice()
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected response received")
	}
}

func TestBlockNumberFetching(t *testing.T) {
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)
	response, err := client.GetBlockNumber()
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected response received")
	}
}
