package eth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	httpProtocol = "http"
	https        = "https"
	ws           = "ws"
	wss          = "wss"
	stdIO        = "stdio"
	ipc          = ""
)

// Client client for connection to the Ethereum.
type Client struct {
	cfg           *Config
	c             *ethclient.Client
	httpTransport *http.Transport
}

// NewClient creates client for connection to the Ethereum.
func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	u, err := url.Parse(cfg.GethURL)
	if err != nil {
		return nil, err
	}

	client := &Client{}

	var rpcClient *rpc.Client

	switch u.Scheme {
	case httpProtocol, https:
		httpTransport := transport(cfg.HttpClient)

		rpcClient, err = rpc.DialHTTPWithClient(cfg.GethURL,
			httpClient(cfg.HttpClient, httpTransport))

		client.httpTransport = httpTransport
	case ws, wss:
		rpcClient, err = rpc.DialWebsocket(ctx, cfg.GethURL, "")
	case stdIO:
		rpcClient, err = rpc.DialStdIO(ctx)
	case ipc:
		rpcClient, err = rpc.DialIPC(ctx, cfg.GethURL)
	default:
		return nil, fmt.Errorf("no known transport"+
			" for URL scheme %q", u.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create"+
			" rpc client %s", err)
	}

	client.c = ethclient.NewClient(rpcClient)
	return client, nil
}

// Close closes client.
func (c *Client) Close() {
	c.c.Close()
}

// CloseIdleConnections closes any connections which were previously
// connected from previous requests.
func (c *Client) CloseIdleConnections() {
	if c.httpTransport != nil {
		c.httpTransport.CloseIdleConnections()
	}
}

// EthClient returns client needed to work with contracts on a read-write basis.
func (c *Client) EthClient() *ethclient.Client {
	return c.c
}

func httpClient(cfg *httpClientConf, transport *http.Transport) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   toTime(cfg.RequestTimeout),
	}
}

func toTime(val uint64) time.Duration {
	return time.Duration(time.Duration(val) * time.Second)
}

func transport(config *httpClientConf) *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: toTime(
				config.DialTimeout),
			DualStack: true,
			KeepAlive: toTime(config.KeepAliveTimeout),
		}).DialContext,
		IdleConnTimeout: toTime(config.IdleConnTimeout),
		TLSHandshakeTimeout: toTime(
			config.TLSHandshakeTimeout),
		ResponseHeaderTimeout: toTime(
			config.ResponseHeaderTimeout),
	}
}
