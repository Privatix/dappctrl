package somcserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// Client can retrieve data from agents somc server.
type Client struct {
	client   *http.Client
	hostname string
}

// NewClient creates a new Client.
func NewClient(client *http.Client, hostname string) *Client {
	return &Client{client, hostname}
}

// GetOffering gets offering message through tor net.
func (c *Client) GetOffering(hash data.HexString) (data.Base64String, error) {
	return c.requestWithPayload("api_offering", string(hash))
}

// GetEndpoint gets endpoint message through tor net.
func (c *Client) GetEndpoint(channelKey data.Base64String) (data.Base64String, error) {
	return c.requestWithPayload("api_endpoint", string(channelKey))
}

func (c *Client) requestWithPayload(method string, param string) (data.Base64String, error) {
	resp, err := c.request(method, param)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return c.extractResult(resp)
}

func (c *Client) request(method string, param string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s/http", c.hostname)
	payload := fmt.Sprintf(`{"method": "%s",
		"params": ["%s"], "id": "%s"}`, method, param, util.NewUUID())
	return c.client.Post(url, "application/json", strings.NewReader(payload))
}

func (c *Client) extractResult(resp *http.Response) (data.Base64String, error) {
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