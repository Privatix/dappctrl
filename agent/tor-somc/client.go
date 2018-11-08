package somc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/privatix/dappctrl/data"
)

// GetOffering gets offering message through tor net.
func GetOffering(hostname string, hash data.Base64String,
	sock uint) (data.Base64String, error) {
	payload := fmt.Sprintf(`{"method": "api_offering",
		"params": ["%s"], "id": 67}`, hash)
	return requestWithPayload(hostname, payload, sock)
}

// GetEndpoint gets endpoint message through tor net.
func GetEndpoint(hostname string, channelKey data.Base64String,
	sock uint) (data.Base64String, error) {
	payload := fmt.Sprintf(`{"method": "api_endpoint",
		"params": ["%s"], "id": 67}`, channelKey)
	return requestWithPayload(hostname, payload, sock)
}

func requestWithPayload(hostname, payload string,
	sock uint) (data.Base64String, error) {
	resp, err := request(hostname, payload, sock)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return extractResult(resp)
}

func request(hostname, payload string, sock uint) (*http.Response, error) {
	url := fmt.Sprintf("http://%s/http", hostname)
	client, err := newClient(sock)
	if err != nil {
		return nil, err
	}
	return client.Post(url, "application/json", strings.NewReader(payload))
}

func extractResult(resp *http.Response) (data.Base64String, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ret struct {
		Result *data.Base64String `json:"result"`
	}

	if err := json.Unmarshal(body, &ret); err != nil {
		return "", err
	}

	if ret.Result == nil {
		return "", fmt.Errorf("unknown reply: %s", body)
	}

	return *ret.Result, nil
}

func newClient(sock uint) (*http.Client, error) {
	torProxyURL, err := url.Parse(fmt.Sprint("socks5://127.0.0.1:", sock))
	if err != nil {
		return nil, err
	}

	// Set up a custom HTTP transport to use the proxy and create the client
	torTransport := &http.Transport{Proxy: http.ProxyURL(torProxyURL)}
	return &http.Client{
		Transport: torTransport,
		Timeout:   time.Second * 30,
	}, nil
}
