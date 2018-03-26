// +build !notest

package data

import (
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"log"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/common/number"
	"github.com/ethereum/go-ethereum/crypto"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

// TestData is a container for testing data items.
type TestData struct {
	Channel  string
	Password string
}

// These are functions for shortening testing boilerplate.

// NewTestDB creates a new database connection.
func NewTestDB(conf *DBConfig, logger *util.Logger) *reform.DB {
	db, err := NewDB(conf, logger)
	if err != nil {
		log.Fatalf("failed to open db connection: %s\n", err)
	}
	return db
}

// NewTestUser returns new subject
func NewTestUser() *User {
	priv, _ := ecdsa.GenerateKey(crypto.S256(), cryptorand.Reader)
	b := crypto.FromECDSA(priv)
	privB64 := FromBytes(b)
	priv, _ = crypto.ToECDSA(b)
	pub := FromBytes(
		crypto.FromECDSAPub(&priv.PublicKey))
	return &User{
		ID:         util.NewUUID(),
		PrivateKey: &privB64,
		PublicKey:  pub,
	}
}

// NewTestProduct returns new product.
func NewTestProduct() *Product {
	return &Product{
		ID:           util.NewUUID(),
		Name:         "Test product",
		UsageRepType: ProductUsageTotal,
	}
}

// NewTestTemplate returns new tempalte.
func NewTestTemplate(kind string) *Template {
	return &Template{
		ID:   util.NewUUID(),
		Raw:  []byte("{}"),
		Kind: kind,
	}
}

// NewTestOffering returns new offering.
func NewTestOffering(agent, product, tpl string) *Offering {
	return &Offering{
		ID:               util.NewUUID(),
		Template:         tpl,
		Agent:            agent,
		Product:          product,
		Supply:           1,
		Status:           MsgChPublished,
		UnitType:         UnitSeconds,
		BillingType:      BillingPostpaid,
		BillingInterval:  100,
		Nonce:            util.NewUUID(),
		AdditionalParams: []byte("{}"),
		SetupPrice:       11,
		UnitPrice:        22,
	}
}

// NewTestChannel returns new channel.
func NewTestChannel(agent, client *User, offering *Offering,
	balance, deposit int64, status string) *Channel {
	return &Channel{
		ID:             util.NewUUID(),
		Agent:          agent.ID,
		Client:         client.ID,
		Offering:       offering.ID,
		Block:          uint(rand.Intn(99999999)),
		ChannelStatus:  status,
		ServiceStatus:  ServiceActive,
		TotalDeposit:   FromBytes(number.Big(deposit).Bytes()),
		ReceiptBalance: FromBytes(number.Big(balance).Bytes()),
	}
}

// NewTestEndpoint returns new endpoint.
func NewTestEndpoint(chanID, tplID string) *Endpoint {
	return &Endpoint{
		ID:               util.NewUUID(),
		Template:         tplID,
		Channel:          chanID,
		Status:           MsgBChainPublished,
		AdditionalParams: []byte("{}"),
	}
}

// NewTestSession returns new session.
func NewTestSession(chanID string) *Session {
	return &Session{
		ID:      util.NewUUID(),
		Channel: chanID,
		Started: time.Now(),
	}
}
