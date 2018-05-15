// +build !notest

package data

import (
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

// TestEncryptedKey is a key encryption simplified for tests performance.
func TestEncryptedKey(pkey *ecdsa.PrivateKey, auth string) (string, error) {
	return FromBytes(crypto.FromECDSA(pkey)) + "AUTH:" + auth, nil
}

// TestToPrivateKey is a key decryption simplified for tests performance.
func TestToPrivateKey(keyB64, auth string) (*ecdsa.PrivateKey, error) {
	split := strings.Split(keyB64, "AUTH:")
	keyB64 = split[0]
	authStored := split[1]
	if auth != authStored {
		return nil, fmt.Errorf("passphrase didn't match")
	}
	keyBytes, err := ToBytes(keyB64)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(keyBytes)
}

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

// NewTestUser returns new subject.
func NewTestUser() *User {
	priv, _ := ecdsa.GenerateKey(crypto.S256(), cryptorand.Reader)
	b := crypto.FromECDSA(priv)
	priv, _ = crypto.ToECDSA(b)
	pub := FromBytes(
		crypto.FromECDSAPub(&priv.PublicKey))
	return &User{
		ID:        util.NewUUID(),
		EthAddr:   util.NewUUID()[:28],
		PublicKey: pub,
	}
}

// NewTestAccount returns new account.
func NewTestAccount(auth string) *Account {
	priv, _ := ecdsa.GenerateKey(crypto.S256(), cryptorand.Reader)
	pub := FromBytes(
		crypto.FromECDSAPub(&priv.PublicKey))
	addr := FromBytes(crypto.PubkeyToAddress(priv.PublicKey).Bytes())
	pkEcnrypted, _ := TestEncryptedKey(priv, auth)
	return &Account{
		ID:         util.NewUUID(),
		EthAddr:    addr,
		PublicKey:  pub,
		PrivateKey: pkEcnrypted,
		IsDefault:  true,
		InUse:      true,
		Name:       util.NewUUID()[:30],
		PTCBalance: 0,
		PSCBalance: 0,
		EthBalance: B64BigInt(FromBytes(big.NewInt(1).Bytes())),
	}
}

// Test authentication constants.
const (
	TestPassword     = "secret"
	TestPasswordHash = "7U9gC4AZsSZ9E8NabVkw8nHRlFCJe0o_Yh9qMlIaGAg="
	TestSalt         = 6012867121110302348
)

// NewTestProduct returns new product.
func NewTestProduct() *Product {
	return &Product{
		ID:           util.NewUUID(),
		Name:         "Test product",
		UsageRepType: ProductUsageTotal,
		Salt:         TestSalt,
		Password:     TestPasswordHash,
		ClientIdent:  ClientIdentByChannelID,
		Config:       []byte("{}"),
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
	offering := &Offering{
		ID:                 util.NewUUID(),
		OfferStatus:        OfferRegister,
		BlockNumberUpdated: 1,
		Template:           tpl,
		Agent:              agent,
		Product:            product,
		Supply:             1,
		Status:             MsgChPublished,
		UnitType:           UnitSeconds,
		BillingType:        BillingPostpaid,
		BillingInterval:    100,
		AdditionalParams:   []byte("{}"),
		SetupPrice:         11,
		UnitPrice:          22,
	}
	offering.Hash = FromBytes(OfferingHash(offering))
	return offering
}

// NewTestChannel returns new channel.
func NewTestChannel(agent, client, offering string,
	balance, deposit uint64, status string) *Channel {
	return &Channel{
		ID:             util.NewUUID(),
		Agent:          agent,
		Client:         client,
		Offering:       offering,
		Block:          uint(rand.Intn(99999999)),
		ChannelStatus:  status,
		ServiceStatus:  ServiceActive,
		TotalDeposit:   deposit,
		ReceiptBalance: balance,
		Salt:           TestSalt,
		Password:       TestPasswordHash,
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

// BeginTestTX begins a test transaction.
func BeginTestTX(t *testing.T, db *reform.DB) *reform.TX {
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %s", err)
	}
	return tx
}

// CommitTestTX commits a test transaction.
func CommitTestTX(t *testing.T, tx *reform.TX) {
	if err := tx.Commit(); err != nil {
		t.Fatalf("failed to commit transaction: %s", err)
	}
}

// RollbackTestTX rollbacks a test transaction.
func RollbackTestTX(t *testing.T, tx *reform.TX) {
	if err := tx.Rollback(); err != nil {
		t.Fatalf("failed to rollback transaction: %s", err)
	}
}

// InsertToTestDB inserts rows to test DB.
func InsertToTestDB(t *testing.T, db *reform.DB, rows ...reform.Struct) {
	tx := BeginTestTX(t, db)
	for _, v := range rows {
		if err := tx.Insert(v); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to insert %T: %s", v, err)
		}
	}
	CommitTestTX(t, tx)
}

// SaveToTestDB saves records to test DB.
func SaveToTestDB(t *testing.T, db *reform.DB, recs ...reform.Record) {
	tx := BeginTestTX(t, db)
	for _, v := range recs {
		if err := tx.Save(v); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to save %T: %s", v, err)
		}
	}
	CommitTestTX(t, tx)
}

// DeleteFromTestDB deletes records from test DB.
func DeleteFromTestDB(t *testing.T, db *reform.DB, recs ...reform.Record) {
	tx := BeginTestTX(t, db)
	for _, v := range recs {
		if err := tx.Delete(v); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to delete %T: %s", v, err)
		}
	}
	CommitTestTX(t, tx)
}

// ReloadFromTestDB reloads records from test DB.
func ReloadFromTestDB(t *testing.T, db *reform.DB, recs ...reform.Record) {
	for _, v := range recs {
		if err := db.Reload(v); err != nil {
			t.Fatalf("failed to reload %T: %s", v, err)
		}
	}
}

// CleanTestDB deletes all records from all test DB tables.
func CleanTestDB(t *testing.T, db *reform.DB) {
	tx := BeginTestTX(t, db)
	for _, v := range []reform.View{JobTable, EndpointTable, SessionTable,
		ChannelTable, OfferingTable, UserTable, AccountTable,
		ProductTable, TemplateTable, ContractTable, SettingTable} {
		if _, err := tx.DeleteFrom(v, ""); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to clean DB: %s", err)
		}
	}
	CommitTestTX(t, tx)
}

type TestFixture struct {
	T        *testing.T
	DB       *reform.DB
	Product  *Product
	Account  *Account
	Template *Template
	Offering *Offering
	Channel  *Channel
}

func NewTestFixture(t *testing.T, db *reform.DB) *TestFixture {
	prod := NewTestProduct()
	acc := NewTestAccount(TestPassword)
	tmpl := NewTestTemplate(TemplateOffer)
	off := NewTestOffering(acc.EthAddr, prod.ID, tmpl.ID)
	ch := NewTestChannel(
		acc.EthAddr, acc.EthAddr, off.ID, 0, 0, ChannelActive)

	InsertToTestDB(t, db, prod, acc, tmpl, off, ch)

	return &TestFixture{
		T:        t,
		DB:       db,
		Product:  prod,
		Account:  acc,
		Template: tmpl,
		Offering: off,
		Channel:  ch,
	}
}

func (f *TestFixture) Close() {
	DeleteFromTestDB(f.T, f.DB, f.Channel,
		f.Offering, f.Template, f.Account, f.Product)
}
