package eth

// This module provides low-level methods for accessing ethereum logs.
// For detailed API description, please refer to:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html

import (
	"encoding/json"
	"errors"
	"fmt"
)

type LogsAPIRecord struct {
	Type                string   `json:"type"`
	TransactionIndexHex string   `json:"transactionIndex"`
	LogIndexHex         string   `json:"logIndex"`
	TransactionHash     string   `json:"transactionHash"`
	Address             string   `json:"address"`
	BlockHash           string   `json:"blockHash"`
	Data                string   `json:"data"`
	Topics              []string `json:"topics"`
	BlockNumberHex      string   `json:"blockNumber"`
}

type LogsAPIResponse struct {
	apiResponse
	Result []LogsAPIRecord `json:"result"`
}

type TopicFilter map[int][]string

func (f TopicFilter) MarshalJSON() ([]byte, error) {
	maxIndex := 0
	for k := range f {
		if k > maxIndex {
			maxIndex = k
		}
	}

	var buf bytes.Buffer
	if _, err := buf.WriteRune('['); err != nil {
		return nil, err
	}

	for i := 0; i <= maxIndex; i++ {
		if i > 0 {
			if _, err := buf.WriteRune(','); err != nil {
				return nil, err
			}
		}

		if b, err := json.Marshal((f)[i]); err == nil {
			if _, err := buf.Write(b); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if _, err := buf.WriteRune(']'); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (f *TopicFilter) ExpectAt(index int, values ...string) {
	if *f == nil {
		*f = make(TopicFilter)
	}
	(*f)[index] = values
}

// GetLog fetches logs form remote geth node.
//
// "topics" contains a topic filter (or nil as a wildcard).
// "fromBlock" - specifies first block number **from** which lookup must be performed.
// "toBlock" - specifies last block number **to** which lookup must be performed.
//
// Tests: logs_test/TestNormalLogsFetching
// Tests: logs_test/TestNegativeLogsFetching
func (e *EthereumClient) GetLogs(contractAddress string, topics TopicFilter, fromBlock, toBlock string) (*LogsAPIResponse, error) {
	if contractAddress == "" {
		return nil, errors.New("contract address is required")
	}

	if fromBlock == "" {
		fromBlock = "earliest"
	}

	if toBlock == "" {
		toBlock = BlockLatest
	}

	topicsJson, err := json.Marshal(topics)
	if err != nil {
		return nil, errors.New("can't marshall topic filter: " + err.Error())
	}

	params := fmt.Sprintf(`{"topics":%s,"address":"%s","fromBlock":"%s","toBlock":"%s"}`,
		topicsJson, contractAddress, fromBlock, toBlock)

	response := &LogsAPIResponse{}
	err = e.fetch("eth_getLogs", params, response)
	if err != nil {
		return nil, errors.New("can't fetch response: " + err.Error())
	}

	return response, nil
}
