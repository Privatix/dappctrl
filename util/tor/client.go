package tor

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// NewHTTPClient returns a client that speaks to tor open sock.
func NewHTTPClient(sock uint) (*http.Client, error) {
	torProxyURL, err := url.Parse(fmt.Sprint("socks5://127.0.0.1:", sock))
	if err != nil {
		return nil, err
	}

	// Set up a custom HTTP transport to use the proxy and create the client
	torTransport := &http.Transport{Proxy: http.ProxyURL(torProxyURL)}
	return &http.Client{
		Transport: torTransport,
		Timeout:   time.Second * 10,
	}, nil
}
