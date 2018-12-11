package somc

import (
	"github.com/privatix/dappctrl/agent/somcserver"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/tor"
)

// Client is expected somc clients interface.
type Client interface {
	Endpoint(data.Base64String) (data.Base64String, error)
	Offering(data.HexString) (data.Base64String, error)
	Ping() error
}

// ClientBuilderInterface an abstract layer for builder, introduced mainly for tests.
type ClientBuilderInterface interface {
	NewClient(uint8, data.Base64String) (Client, error)
}

// Contracts. Clients must implement interface.
var _ Client = new(somcserver.Client)

// ClientBuilder responsible for creating Client's.
type ClientBuilder struct {
	torSocks uint
}

// NewClientBuilder creates new ClientBuilder.
func NewClientBuilder(torSocks uint) *ClientBuilder {
	return &ClientBuilder{torSocks}
}

// NewClient returns new client instance based given somc type and data.
func (b *ClientBuilder) NewClient(somcType uint8, somcData data.Base64String) (Client, error) {
	if somcType == data.OfferingSOMCTor {
		hostnameBytes, err := data.ToBytes(somcData)
		if err != nil {
			return nil, err
		}
		torClient, err := tor.NewHTTPClient(b.torSocks)
		return somcserver.NewClient(torClient, string(hostnameBytes)), nil
	}

	return nil, ErrUnknownSOMCType
}
