package sess

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/privatix/dappctrl/data"
)

// Client is an adapter on top of rpc client to talk to session server.
type Client struct {
	client   *rpc.Client
	product  string
	password string
}

// Dial returns client with established connection to session server.
func Dial(ctx context.Context, endpoint, origin, product, password string) (*Client, error) {
	c, err := rpc.DialWebsocket(ctx, endpoint, origin)
	if err != nil {
		return nil, err
	}
	return &Client{c, product, password}, nil
}

// ConnChange subscribes to connection changes.
func (c *Client) ConnChange(connCh chan *ConnChangeResult) (*rpc.ClientSubscription, error) {
	return c.client.Subscribe(context.Background(), "sess", connCh,
		"connChange", c.product, c.password)
}

// GetEndpoint finds endpoint by key.
func (c *Client) GetEndpoint(key string) (*data.Endpoint, error) {
	var endpoint data.Endpoint
	err := c.callSess(&endpoint, "getEndpoint", key)
	if err != nil {
		return nil, err
	}

	return &endpoint, nil
}

// ServiceReady sends signal that service is ready.
func (c *Client) ServiceReady(key string) error {
	return c.callSess(nil, "serviceReady", key)
}

// AuthClient verifies user credentials.
func (c *Client) AuthClient(user, pass string) error {
	return c.callSess(nil, "authClient", user, pass)
}

// StartSession start session.
func (c *Client) StartSession(trustedIP, key string, port uint16) (*data.Offering, error) {
	var offer data.Offering

	err := c.callSess(&offer, "startSession", key, trustedIP, port)
	if err != nil {
		return nil, err
	}

	return &offer, nil
}

// UpdateSession updates session.
func (c *Client) UpdateSession(key string, units uint64) error {
	return c.callSess(nil, "updateSession", key, units)
}

// StopSession updates session.
func (c *Client) StopSession(key string) error {
	return c.callSess(nil, "stopSession", key)
}

// SetProductConfig sets product config.
func (c *Client) SetProductConfig(config map[string]string) error {
	return c.callSess(nil, "setProductConfig", config)
}

func (c *Client) callSess(result interface{}, method string, args ...interface{}) error {
	creds := []interface{}{c.product, c.password}
	return c.client.Call(result, "sess_"+method, append(creds, args...)...)
}
