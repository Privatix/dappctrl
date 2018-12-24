package somc

import (
	"net/http"
	"sort"

	"github.com/privatix/dappctrl/agent/somcsrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/tor"
	reform "gopkg.in/reform.v1"
)

// Client is expected somc clients interface.
type Client interface {
	Endpoint(data.Base64String) (data.Base64String, error)
	Offering(data.HexString) (data.Base64String, error)
	Ping() error
}

// ClientBuilderInterface an abstract layer for builder, introduced mainly for tests.
type ClientBuilderInterface interface {
	NewClient(uint8, string) (Client, error)
}

// Contracts. Clients must implement interface.
var _ Client = new(somcsrv.Client)

// ClientBuilder responsible for creating Client's.
type ClientBuilder struct {
	torConf *TorClientConfig
	db      *reform.DB
}

// NewClientBuilder creates new ClientBuilder.
func NewClientBuilder(conf *TorClientConfig, db *reform.DB) *ClientBuilder {
	return &ClientBuilder{conf, db}
}

type somcProps struct {
	code uint8
	data data.Base64String
}

type byCode []somcProps

func (c byCode) Len() int           { return len(c) }
func (c byCode) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c byCode) Less(i, j int) bool { return c[i].code < c[j].code }

// NewClient returns new client instance based given somc type and data.
func (b *ClientBuilder) NewClient(somcType uint8, somcData string) (Client, error) {
	transports := b.extractTransports(somcType, somcData)

	usingTor, err := b.canUseTor()
	if err != nil {
		return nil, err
	}

	usingDirect, err := b.canUseDirect()
	if err != nil {
		return nil, err
	}

	for _, transport := range transports {
		dataBytes, err := data.ToBytes(transport.data)
		if err != nil {
			return nil, err
		}

		if transport.code == torCode && usingTor {
			return b.torClient(dataBytes)
		}

		if transport.code == directCode && usingDirect {
			return b.directClient(dataBytes)
		}
	}

	return nil, ErrUnknownSOMCType
}

func (b *ClientBuilder) canUseTor() (bool, error) {
	return data.ReadBoolSetting(b.db.Querier, data.SettingSOMCTOR)
}

func (b *ClientBuilder) canUseDirect() (bool, error) {
	return data.ReadBoolSetting(b.db.Querier, data.SettingSOMCDirect)
}

func (b *ClientBuilder) extractTransports(somcType uint8, somcData string) []somcProps {
	transports := []somcProps{}
	dataParts := extractURLBase64Strings(somcData)

	// Higher codes must be extracted first. But priority goes from less to high.
	// To deal with this, we extract everything and sort by code.
	for _, code := range []uint8{directCode, torCode} {
		if somcType >= code {
			somcType -= code
			transports = append(transports, somcProps{code, dataParts[0]})
			dataParts = dataParts[1:]
		}
	}

	// Priority goes by code. The less code's value the higher it's priority.
	sort.Sort(byCode(transports))

	return transports
}

func (b *ClientBuilder) torClient(dataBytes []byte) (Client, error) {
	c, err := tor.NewHTTPClient(b.torConf.Socks)
	if err != nil {
		return nil, err
	}

	return somcsrv.NewClient(c, string(dataBytes)), nil
}

func (b *ClientBuilder) directClient(dataBytes []byte) (Client, error) {
	c := &http.Client{}
	return somcsrv.NewClient(c, string(dataBytes)), nil
}
