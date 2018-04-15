package lib

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

type BlockNumberAPIResponse GasPriceAPIResponse

// GetBlockNumber returns the number of most recent block in blockchain.
// For the details, please, refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html#eth_blocknumber
func (e *EthereumClient) GetBlockNumber() (*BlockNumberAPIResponse, error) {
	response := &BlockNumberAPIResponse{}
	return response, e.fetch("eth_blockNumber", "", response)
}

type BalanceAPIResponse GasPriceAPIResponse

// GetBalance returns the balance of the account of given address in wei.
// For the details, please, refer to:
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getbalance
func (e *EthereumClient) GetBalance(addressHex, blockNumberHex string) (*BalanceAPIResponse, error) {
	response := &BalanceAPIResponse{}
	return response, e.fetch("eth_getBalance", `"`+
		addressHex+`", "`+
		blockNumberHex+`"`, response)
}

type TransactionReceiptAPIResponse struct {
	apiResponse
	Result struct {
		TransactionHash   string   `json:"transactionHash"`
		TransactionIndex  string   `json:"transactionIndex"`
		BlockHash         string   `json:"blockHash"`
		BlockNumber       string   `json:"blockNumber"`
		GasUsed           string   `json:"gasUsed"`
		CumulativeGasUsed string   `json:"cumulativeGasUsed"`
		ContractAddress   string   `json:"contractAddress"`
		Logs              []string `json:"logs"`
		Status            string   `json:"status"`
		LogsBloom         string   `json:"logsBloom"`
	} `json:"result"`
}

// GetTransactionReceipt returns receipt of the transaction,
// specified by the hash.
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionreceipt
func (e *EthereumClient) GetTransactionReceipt(hash string) (*TransactionReceiptAPIResponse, error) {
	response := &TransactionReceiptAPIResponse{}
	return response, e.fetch("eth_getTransactionReceipt", `"`+hash+`"`, response)
}

type TransactionByHashAPIResponse struct {
	apiResponse
	Result struct {
		Gas              string `json:"gas"`
		GasPrice         string `json:"gasPrice"`
		Hash             string `json:"hash"`
		BlockNumber      string `json:"blockNumber"`
		Value            string `json:"value"`
		From             string `json:"from"`
		Nonce            string `json:"nonce"`
		BlockHash        string `json:"blockHash"`
		Input            string `json:"input"`
		To               string `json:"to"`
		TransactionIndex string `json:"transactionIndex"`
	} `json:"result"`
}

// GetTransactionByHash returns the information about a transaction requested by transaction hash.
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionbyhash
func (e *EthereumClient) GetTransactionByHash(hash string) (*TransactionByHashAPIResponse, error) {
	response := &TransactionByHashAPIResponse{}
	return response, e.fetch("eth_getTransactionByHash", `"`+hash+`"`, response)
}
