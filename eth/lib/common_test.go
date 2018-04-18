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

func getClient() *EthereumClient {
	node := tests.GethEthereumConfig().Geth
	return NewEthereumClient(node.Host, node.Port)
}

func TestGasPriceFetching(t *testing.T) {
	response, err := getClient().GetGasPrice()
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected response received")
	}
}

func TestBlockNumberFetching(t *testing.T) {
	response, err := getClient().GetBlockNumber()
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected response received")
	}
}

func TestTransactionReceipt(t *testing.T) {
	transactionHash := getContractTransactionHash()
	response, err := getClient().GetTransactionReceipt(transactionHash)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result.TransactionHash != transactionHash {
		t.Fatal("Transaction hashes are not equal.")
	}
}

func TestTransactionByHash(t *testing.T) {
	transactionHash := getContractTransactionHash()
	response, err := getClient().GetTransactionByHash(transactionHash)
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

func TestBalanceOnLastBlock(t *testing.T) {
	response, err := getClient().GetBalance(getTestAccountAddress(), "latest")
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected balance occurred")
	}
}

func TestBalanceOnFirstBlock(t *testing.T) {
	response, err := getClient().GetBalance(getTestAccountAddress(), "earliest")
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected balance occurred")
	}
}

func getTestAccountAddress() string {
	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getKeys")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch private key. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make([]interface{}, 0, 0)
	json.Unmarshal(body, &data)

	return data[0].(map[string]interface{})["account"].(string)
}
