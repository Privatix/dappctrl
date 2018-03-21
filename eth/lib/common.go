package lib

import (
	"strings"
)

// This module provides low-level methods for accessing common ethereum info.
// For detailed API description, please refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html

type GasPriceAPIResponse struct {
	apiResponse
	Result string `json:"result"`
}

// GetGasPrice returns current gas price in wei.
// For the details, please, refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html#eth_gasprice
func (e *EthereumClient) GetGasPrice() (*GasPriceAPIResponse, error) {
	response := &GasPriceAPIResponse{}
	return response, e.fetch("eth_gasPrice", "", response)
}

//---------------------------------------------------------------------------------------------------------------------

type BlockNumberAPIResponse GasPriceAPIResponse

// GetBlockNumber returns the number of most recent block in blockchain.
// For the details, please, refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html#eth_blocknumber
func (e *EthereumClient) GetBlockNumber() (*BlockNumberAPIResponse, error) {
	response := &BlockNumberAPIResponse{}
	return response, e.fetch("eth_blockNumber", "", response)
}

//---------------------------------------------------------------------------------------------------------------------

type TransactionReceiptAPIResponse struct {
	apiResponse
	Result string `json:"result"`
}
// GetTransactionReceipt returns receipt of the transaction,
// specified by the hash.
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionreceipt
func (e *EthereumClient) GetTransactionReceipt(hash string) (*TransactionReceiptAPIResponse, error) {
	if len(hash) > 2 {
		if strings.ToLower(hash[:2]) != "0x" {
			hash = "0x" + hash
		}
	}

	response := &TransactionReceiptAPIResponse{}
	return response, e.fetch("eth_getTransactionReceipt", `"` + hash + `"`, response)
}