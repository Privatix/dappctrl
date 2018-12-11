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

// Offering gets offering message through tor net.
func (c *Client) Offering(hash data.HexString) (data.Base64String, error) {
	return c.requestWithPayload("api_offering", string(hash))
}

// Endpoint gets endpoint message through tor net.
func (c *Client) Endpoint(channelKey data.Base64String) (data.Base64String, error) {
	return c.requestWithPayload("api_endpoint", string(channelKey))
}

// Ping returns an error if remote enpoint cannot be reached.
func (c *Client) Ping() error {
	_, err := c.client.Head(c.url())
	return err
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
	payload := fmt.Sprintf(`{"method": "%s",
		"params": ["%s"], "id": "%s"}`, method, param, util.NewUUID())
	return c.client.Post(c.url(), "application/json", strings.NewReader(payload))
}

func (c *Client) url() string {
	return fmt.Sprintf("http://%s/http", c.hostname)
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
