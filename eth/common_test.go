// +build !noethtest

package eth

import (
	"testing"
)

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

func TestTxReceipt(t *testing.T) {
	transactionHash := testTruffleAPI.GetContractTransactionHash()
	response, err := getClient().GetTransactionReceipt(transactionHash)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result.TransactionHash != transactionHash {
		t.Fatal("Transaction hashes are not equal.")
	}
}

func TestTxByHash(t *testing.T) {
	transactionHash := testTruffleAPI.GetContractTransactionHash()
	response, err := getClient().GetTransactionByHash(transactionHash)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result.Hash != transactionHash {
		t.Fatal("Transaction hashes are not equal.")
	}
}

func TestBalanceOnLastBlock(t *testing.T) {
	response, err := getClient().GetBalance(testTruffleAPI.GetTestAccountAddress(), BlockLatest)
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected balance occurred")
	}
}

func TestBalanceOnFirstBlock(t *testing.T) {
	response, err := getClient().GetBalance(testTruffleAPI.GetTestAccountAddress(), "earliest")
	if err != nil {
		t.Fatal(err)
	}

	if response.Result == "" {
		t.Fatal("Unexpected balance occurred")
	}
}
