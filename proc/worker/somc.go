package worker

import (
	"github.com/privatix/dappctrl/agent/somcserver"
	"github.com/privatix/dappctrl/util/tor"

	"github.com/privatix/dappctrl/data"
)

// SOMCClient is expected interface for somc clients used by worker package.
type SOMCClient interface {
	Endpoint(data.Base64String) (data.Base64String, error)
	Offering(data.HexString) (data.Base64String, error)
}

// SOMCClientBuilderInterface an abstract layer for builder, introduced mainly for tests.
type SOMCClientBuilderInterface interface {
	NewClient(uint8, data.Base64String) (SOMCClient, error)
}

// Contracts.
var _ SOMCClient = new(somcserver.Client)

// SOMCClientBuilder responsible for creating SOMCClient's.
type SOMCClientBuilder struct {
	torSocks uint
}

// NewSOMCClientBuilder creates new SOMCClientBuilder.
func NewSOMCClientBuilder(torSocks uint) *SOMCClientBuilder {
	return &SOMCClientBuilder{torSocks}
}

// NewClient returns new client instance based given somc type and data.
func (b *SOMCClientBuilder) NewClient(somcType uint8, somcData data.Base64String) (SOMCClient, error) {
	if somcType == data.OfferingSOMCTor {
		hostnameBytes, err := data.ToBytes(somcData)
		if err != nil {
			return nil, err
		}
		torClient, err := tor.NewHTTPClient(b.torSocks)
		return somcserver.NewClient(torClient, string(hostnameBytes)), nil
	}

	return nil, ErrUnknownSomcType
}
