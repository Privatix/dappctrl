package eth

// Implements client for ethereum network.
//
// For details about contracts methods calling:
// https://github.com/ethereum/go-ethereum/wiki/Native-DApps:-Go-bindings-to-Ethereum-contracts
//
// Note:
// "abigen_linux" is available in contract/tools, no need for building it itself.
//
// For details about events (logs in EthereumConf terminology) fetching:
// https://ethereumbuilders.gitbooks.io/guide/content/en/ethereum_json_rpc.html

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Block labels.
const (
	BlockLatest          = "latest"
	sessionCacheCapacity = 1024
)

var (
	tlsConfig = &tls.Config{
		ClientSessionCache: tls.NewLRUClientSessionCache(
			sessionCacheCapacity),
	}
)

// Config is a configuration for Ethereum client.
type Config struct {
	Contract contract
	GethURL  string
	Timeout  timeout
}

type contract struct {
	PTCAddrHex string
	PSCAddrHex string
}

type timeout struct {
	ResponseHeaderTimeout uint
	TLSHandshakeTimeout   uint
	ExpectContinueTimeout uint
}

// EthereumClient implementation of client logic for the ethereum geth node.
// Uses JSON RPC API of geth for communication with remote node.
type EthereumClient struct {
	GethURL string

	client    http.Client
	requestID uint64
}

// NewConfig creates a default Ethereum client configuration.
func NewConfig() *Config {
	return &Config{
		Timeout: timeout{
			ResponseHeaderTimeout: 20,
			TLSHandshakeTimeout:   5,
			ExpectContinueTimeout: 5,
		},
	}
}

// NewEtherClient make Ethereum client.
func NewEtherClient(conf *Config) (*ethclient.Client, error) {
	u, err := url.Parse(conf.GethURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ethclient.Dial(conf.GethURL)
	}
	httpClient := newHTTPClient(conf)
	httpEthClient, err := rpc.DialHTTPWithClient(conf.GethURL, httpClient)
	if err != nil {
		return nil, err
	}

	return ethclient.NewClient(httpEthClient), nil
}

func newHTTPClient(conf *Config) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: tlsConfig,
			ResponseHeaderTimeout: time.Duration(
				conf.Timeout.ResponseHeaderTimeout) * time.Second,
			TLSHandshakeTimeout: time.Duration(
				conf.Timeout.TLSHandshakeTimeout) * time.Second,
			ExpectContinueTimeout: time.Duration(
				conf.Timeout.ExpectContinueTimeout) * time.Second,
		},
		Timeout: time.Duration(
			conf.Timeout.ResponseHeaderTimeout) * time.Second,
	}
}

// NewEthereumClient creates and returns instance of client for remote ethereum node,
// that is available via specified host and port.
func NewEthereumClient(gethURL string) *EthereumClient {
	return &EthereumClient{
		GethURL: gethURL,

		// By default, standard http-client does not uses any timeout for its operations.
		// But, there is non zero probability, that remote geth-node would hang for a long time.
		// To avoid cascade client/agent side application hangs - custom timeout is used.
		client: http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// apiResponse is a base geth API response.
type apiResponse struct {
	ID      uint64 `json:"id"`
	JSONRPC string `json:"jsonrpc"`

	// All responses also contain "result" field,
	// but from method to method it might have various different types,
	// so it is delegated to the specified response types.
}

// Fills common parameters "method" and "params",
// and calls json-RPC API of the remote geth-node.
// In case of success - tries to unmarshal received data
// to the appropriate structure type ("result" argument).
//
// Tests: this is a base method for all raw API calls
// so, it is automatically covered by the all tests of all low-level methods,
// for example, GetBlockNumber()
func (e *EthereumClient) fetch(method, params string, result interface{}) error {
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":[%s]}`, method, params)
	httpResponse, err := e.client.Post(e.GethURL, "application/json", strings.NewReader(body))
	if err != nil {
		return errors.New("can't do API call: " + err.Error())
	}

	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return errors.New("can't do API call. API responded with code: " +
			fmt.Sprint(httpResponse.StatusCode))
	}

	response, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return errors.New("can't read response data: " + err.Error())
	}

	err = json.Unmarshal(response, result)
	if err != nil {
		return errors.New("can't unmarshal response data: " + err.Error())
	}

	return nil
}

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
