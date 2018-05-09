package eth

const (
	BlockLatest = "latest"
)

// BlockNumberAPIResponse implements wrapper for ethereum JSON RPC API response.
// Please see corresponding web3.js method for the details.
type BlockNumberAPIResponse struct {
	apiResponse
	Result string `json:"result"`
}

// GetBlockNumber returns the number of most recent block in blockchain.
// For the details, please, refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html#eth_blocknumber
func (e *EthereumClient) GetBlockNumber() (*BlockNumberAPIResponse, error) {
	response := &BlockNumberAPIResponse{}
	return response, e.fetch("eth_blockNumber", "", response)
}
