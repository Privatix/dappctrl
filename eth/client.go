package eth

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/privatix/dappctrl/util/log"
)

const (
	httpProtocol = "http"
	https        = "https"
	ws           = "ws"
	wss          = "wss"
	stdIO        = "stdio"
	ipc          = ""
)

// client client for connection to the Ethereum.
type client struct {
	cancel        context.CancelFunc
	cfg           *Config
	logger        log.Logger
	client        *ethclient.Client
	httpTransport *http.Transport
}

// newClient creates client for connection to the Ethereum.
func newClient(cfg *Config, logger log.Logger) (*client, error) {
	logger2 := logger.Add("method", "newClient", "config", cfg)

	u, err := url.Parse(cfg.GethURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &client{cancel: cancel, cfg: cfg, logger: logger}

	var rpcClient *rpc.Client

	switch u.Scheme {
	case httpProtocol, https:
		httpTransport := transport(cfg.HTTPClient)

		rpcClient, err = rpc.DialHTTPWithClient(cfg.GethURL,
			httpClient(cfg.HTTPClient, httpTransport))

		c.httpTransport = httpTransport
	case ws, wss:
		rpcClient, err = rpc.DialWebsocket(ctx, cfg.GethURL, "")
	case stdIO:
		rpcClient, err = rpc.DialStdIO(ctx)
	case ipc:
		rpcClient, err = rpc.DialIPC(ctx, cfg.GethURL)
	default:
		logger2.Add("scheme", u.Scheme).Error(err.Error())
		return nil, ErrURLScheme
	}
	if err != nil {
		logger2.Error(err.Error())
		return nil, ErrCreateClient
	}

	c.client = ethclient.NewClient(rpcClient)
	return c, nil
}

// close closes client.
func (c *client) close() {
	c.client.Close()
}

// closeIdleConnections closes any connections which were previously
// connected from previous requests.
func (c *client) closeIdleConnections() {
	if c.httpTransport != nil {
		c.httpTransport.CloseIdleConnections()
	}
}

// ethClient returns client needed to work with contracts on a read-write basis.
func (c *client) ethClient() *ethclient.Client {
	return c.client
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
