// +build !noethtest

package lib

import (
	"encoding/json"
	"github.com/privatix/dappctrl/eth/lib/tests"
	"io/ioutil"
	"log"
	"net/http"
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

func TestTransactionReceipt(t *testing.T) {
	transactionHash := getContractTransactionHash()
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)
	response, err := client.GetTransactionReceipt(transactionHash)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result.TransactionHash != transactionHash {
		t.Fatal("Transaction hashes are not equal.")
	}
}

func TestTransactionByHash(t *testing.T) {
	transactionHash := getContractTransactionHash()
	node := tests.GethEthereumConfig().Geth
	client := NewEthereumClient(node.Host, node.Port)
	response, err := client.GetTransactionByHash(transactionHash)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result.Hash != transactionHash {
		t.Fatal("Transaction hashes are not equal.")
	}
}

func getContractTransactionHash() string {
	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	apiResponse, err := http.Get(truffleAPI.Interface() + "/getPSC")
	if err != nil || apiResponse.StatusCode != 200 {
		log.Fatal("Can't fetch PSC address. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(apiResponse.Body)
	defer apiResponse.Body.Close()
	if err != nil {
		log.Fatal("Can't read apiResponse body. It seems that test environment is broken.")
	}

	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	return data["contract"].(map[string]interface{})["transactionHash"].(string)
}
