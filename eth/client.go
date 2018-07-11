package eth

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
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
	Contract struct {
		PTCAddrHex string
		PSCAddrHex string
	}
	GethURL string
	Timeout timeout
}

type timeout struct {
	ResponseHeaderTimeout uint
	TLSHandshakeTimeout   uint
	ExpectContinueTimeout uint
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
