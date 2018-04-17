// +build !noethtest

package lib

import (
	"testing"

	"github.com/privatix/dappctrl/eth/lib/tests"
)

func TestNegativeLogsFetching(t *testing.T) {
	failIfNoError := func(err error, args ...interface{}) {
		if err == nil {
			t.Fatal(args)
		}
	}

	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)

	_, err := client.GetLogs("", []string{"0x0"}, "", "")
	failIfNoError(err, "Error must be returned")

	_, err = client.GetLogs(fetchPSCAddress(), []string{"0x0"}, "", "")
	failIfNoError(err, "Error must be returned")

	_, err = client.GetLogs(fetchPSCAddress(), []string{"", ""}, "", "")
	failIfNoError(err, "Error must be returned")
}

func TestLogsFetchingWithBrokenNetwork(t *testing.T) {
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port+1) // Note: invalid port is used.

	{
		_, err := client.GetLogs(fetchPSCAddress(), []string{EthOfferingCreated}, "", "")
		if err == nil {
			t.Fatal("Error must be returned")
		}
	}
}
