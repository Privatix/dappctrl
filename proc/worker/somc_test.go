package worker

import "github.com/privatix/dappctrl/data"

type testSOMCClient struct {
	v data.Base64String
}

var _ SOMCClient = new(testSOMCClient)

func newTestSOMCClient() *testSOMCClient {
	return &testSOMCClient{}
}

func (c *testSOMCClient) Endpoint(data.Base64String) (data.Base64String, error) {
	return c.v, nil
}

func (c *testSOMCClient) Offering(data.HexString) (data.Base64String, error) {
	return c.v, nil
}

type testSOMCClientBuilder struct {
	c SOMCClient
}

func newTestSOMCClientBuilder(client SOMCClient) *testSOMCClientBuilder {
	return &testSOMCClientBuilder{client}
}

func (b *testSOMCClientBuilder) NewClient(uint8, data.Base64String) (SOMCClient, error) {
	return b.c, nil
}
