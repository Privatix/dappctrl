// +build !notest

package somc

import (
	"github.com/privatix/dappctrl/data"
)

// TestClient for tests.
type TestClient struct {
	V   data.Base64String
	Err error
}

// Must implement interface.
var _ Client = new(TestClient)

// NewTestClient creates new test client instance.
func NewTestClient() *TestClient {
	return &TestClient{}
}

// Endpoint return stored V, Err values.
func (c *TestClient) Endpoint(data.Base64String) (data.Base64String, error) {
	return c.V, c.Err
}

// Offering return stored V, Err values.
func (c *TestClient) Offering(data.HexString) (data.Base64String, error) {
	return c.V, c.Err
}

// Ping return stored Err value.
func (c *TestClient) Ping() error {
	return c.Err
}

// TestClientBuilder for tests.
type TestClientBuilder struct {
	c Client
}

// Must implement interface.
var _ ClientBuilderInterface = new(TestClientBuilder)

// NewTestClientBuilder creates client builder that always returns same client.
func NewTestClientBuilder(client Client) *TestClientBuilder {
	return &TestClientBuilder{client}
}

// NewClient returns stored client instance.
func (b *TestClientBuilder) NewClient(uint8, data.Base64String) (Client, error) {
	return b.c, nil
}
