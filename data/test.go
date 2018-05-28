// +build !notest

package data

import (
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/eth/truffle"
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

// TestToBytes returns binary representation of base64 encoded string or fails.
func TestToBytes(t *testing.T, s string) []byte {
	b, err := base64.URLEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		t.Fatal("failed to decode: ", err)
	}
	return b
}

// TestToHash decodes to hash or fails.
func TestToHash(t *testing.T, h string) common.Hash {
	ret, err := ToHash(h)
	if err != nil {
		t.Fatal("failed to make hash: ", err)
	}
	return ret
}

// TestToAddress decodes to address or fails.
func TestToAddress(t *testing.T, addr string) common.Address {
	ret, err := ToAddress(addr)
	if err != nil {
		t.Fatal("failed to make addr")
	}
	return ret
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
	pub := FromBytes(
		crypto.FromECDSAPub(&priv.PublicKey))
	addr := FromBytes(crypto.PubkeyToAddress(priv.PublicKey).Bytes())
	return &User{
		ID:        util.NewUUID(),
		EthAddr:   addr,
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

// NewEthTestAccount returns new account based on truffle.TestAccount.
func NewEthTestAccount(auth string, acc *truffle.TestAccount) *Account {
	pub := FromBytes(crypto.FromECDSAPub(&acc.PrivateKey.PublicKey))
	addr := FromBytes(acc.Address.Bytes())
	pkEcnrypted, _ := TestEncryptedKey(acc.PrivateKey, auth)
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
	TestPasswordHash = "JDJhJDEwJHNVbWNtTkVwQk5DMkwuOC5OL1BXU08uYkJMMkxjcmthTW1BZklOTUNjNWZDdWNUOU54Tzlp"
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
	tmpl := &Template{
		ID:   util.NewUUID(),
		Raw:  []byte("{}"),
		Kind: kind,
	}
	tmpl.Hash = FromBytes(crypto.Keccak256(tmpl.Raw))
	return tmpl
}

// NewTestOffering returns new offering.
func NewTestOffering(agent, product, tpl string) *Offering {
	fakeMsg := []byte(util.NewUUID())
	offering := &Offering{
		ID:                 util.NewUUID(),
		OfferStatus:        OfferEmpty,
		BlockNumberUpdated: 1,
		Template:           tpl,
		Agent:              agent,
		Hash:               FromBytes(crypto.Keccak256(fakeMsg)),
		Product:            product,
		Supply:             1,
		Status:             MsgUnpublished,
		UnitType:           UnitSeconds,
		BillingType:        BillingPostpaid,
		BillingInterval:    100,
		AdditionalParams:   []byte("{}"),
		SetupPrice:         11,
		UnitPrice:          22,
	}
	return offering
}

// NewTestChannel returns new channel.
func NewTestChannel(agent, client, offering string,
	balance, deposit uint64, status string) *Channel {
	receiptSigFake := FromBytes([]byte("fake-sig"))
	return &Channel{
		ID:               util.NewUUID(),
		Agent:            agent,
		Client:           client,
		Offering:         offering,
		Block:            uint32(rand.Int31()),
		ChannelStatus:    status,
		ServiceStatus:    ServicePending,
		TotalDeposit:     deposit,
		ReceiptBalance:   balance,
		ReceiptSignature: &receiptSigFake,
		Salt:             TestSalt,
		Password:         TestPasswordHash,
	}
}

// NewTestEndpoint returns new endpoint.
func NewTestEndpoint(chanID, tplID string) *Endpoint {
	addr := "addr"
	username := "username"
	password := "password"
	return &Endpoint{
		ID:                     util.NewUUID(),
		Template:               tplID,
		Channel:                chanID,
		Status:                 MsgBChainPublished,
		RawMsg:                 FromBytes([]byte("the message")),
		AdditionalParams:       []byte("{}"),
		PaymentReceiverAddress: &addr,
		ServiceEndpointAddress: &addr,
		Username:               &username,
		Password:               &password,
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

// NewTestJob returns a default test job.
func NewTestJob(jobType, createdBy, relType string) *Job {
	return &Job{
		ID:          util.NewUUID(),
		Status:      JobActive,
		Type:        jobType,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		NotBefore:   time.Now(),
		RelatedType: relType,
		TryCount:    10,
		Data:        []byte("{}"),
	}
}

// NewTestEthLog returns a default test eth log.
func NewTestEthLog() *EthLog {
	return &EthLog{
		ID:          util.NewUUID(),
		TxStatus:    TxMined,
		BlockNumber: uint64(rand.Intn(9999)),
		Topics:      LogTopics{},
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

// FindInTestDB selects a record from test db
func FindInTestDB(t *testing.T, db *reform.DB,
	str reform.Struct, column string, arg interface{}) {
	if err := db.FindOneTo(str, column, arg); err != nil {
		t.Fatalf("failed to find %T: %v (%s)", str, err, util.Caller())
	}
}

// InsertToTestDB inserts rows to test DB.
func InsertToTestDB(t *testing.T, db *reform.DB, rows ...reform.Struct) {
	tx := BeginTestTX(t, db)
	for _, v := range rows {
		if err := tx.Insert(v); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to insert %T: %v (%s)", v, err,
				util.Caller())
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
			t.Fatalf("failed to save %T: %v (%s)", v, err,
				util.Caller())
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
			t.Fatalf("failed to delete %T: %v (%s)", v, err,
				util.Caller())
		}
	}
	CommitTestTX(t, tx)
}

// ReloadFromTestDB reloads records from test DB.
func ReloadFromTestDB(t *testing.T, db *reform.DB, recs ...reform.Record) {
	for _, v := range recs {
		if err := db.Reload(v); err != nil {
			t.Fatalf("failed to reload %T: %v (%s)", v, err,
				util.Caller())
		}
	}
}

// CleanTestDB deletes all records from all test DB tables.
func CleanTestDB(t *testing.T, db *reform.DB) {
	tx := BeginTestTX(t, db)
	for _, v := range []reform.View{EthTxTable, EthLogTable, JobTable,
		EndpointTable, SessionTable, ChannelTable, OfferingTable,
		UserTable, AccountTable, ProductTable, TemplateTable,
		ContractTable, SettingTable} {
		if _, err := tx.DeleteFrom(v, ""); err != nil {
			RollbackTestTX(t, tx)
			t.Fatalf("failed to clean DB: %s", err)
		}
	}
	CommitTestTX(t, tx)
}

// CleanTestTable deletes all records from a given DB table.
func CleanTestTable(t *testing.T, db *reform.DB, tbl reform.Table) {
	if _, err := db.DeleteFrom(tbl, ""); err != nil {
		t.Fatalf("failed to clean %T table: %s", tbl, err)
	}
}

// TestFixture encapsulates a typical set of DB objects useful for testing.
type TestFixture struct {
	T              *testing.T
	DB             *reform.DB
	Product        *Product
	Account        *Account
	User           *User
	TemplateOffer  *Template
	TemplateAccess *Template
	Offering       *Offering
	Channel        *Channel
	Endpoint       *Endpoint
}

// NewTestFixture creates a new test fixture.
func NewTestFixture(t *testing.T, db *reform.DB) *TestFixture {
	prod := NewTestProduct()
	acc := NewTestAccount(TestPassword)
	user := NewTestUser()
	tmpl := NewTestTemplate(TemplateOffer)
	off := NewTestOffering(acc.EthAddr, prod.ID, tmpl.ID)
	ch := NewTestChannel(
		acc.EthAddr, user.EthAddr, off.ID, 0, 0, ChannelActive)
	endpTmpl := NewTestTemplate(TemplateAccess)
	prod.OfferAccessID = &endpTmpl.ID
	endp := NewTestEndpoint(ch.ID, endpTmpl.ID)

	InsertToTestDB(t, db, endpTmpl, prod, acc, user, tmpl, off, ch, endp)

	return &TestFixture{
		T:              t,
		DB:             db,
		Product:        prod,
		Account:        acc,
		User:           user,
		TemplateOffer:  tmpl,
		TemplateAccess: endpTmpl,
		Offering:       off,
		Channel:        ch,
		Endpoint:       endp,
	}
}

func NewEthTestFixture(t *testing.T, db *reform.DB,
	account *truffle.TestAccount) *TestFixture {
	prod := NewTestProduct()
	acc := NewEthTestAccount(TestPassword, account)
	user := NewTestUser()
	tmpl := NewTestTemplate(TemplateOffer)
	off := NewTestOffering(acc.EthAddr, prod.ID, tmpl.ID)
	ch := NewTestChannel(
		acc.EthAddr, user.EthAddr, off.ID, 0, 0, ChannelActive)
	endpTmpl := NewTestTemplate(TemplateAccess)
	endp := NewTestEndpoint(ch.ID, endpTmpl.ID)

	InsertToTestDB(t, db, prod, acc, user, tmpl, off, ch, endpTmpl, endp)

	return &TestFixture{
		T:              t,
		DB:             db,
		Product:        prod,
		Account:        acc,
		User:           user,
		TemplateOffer:  tmpl,
		TemplateAccess: endpTmpl,
		Offering:       off,
		Channel:        ch,
		Endpoint:       endp,
	}
}

// Close closes a given test fixture.
func (f *TestFixture) Close() {
	// (t, db, endpTmpl, prod, acc, user, tmpl, off, ch, endp)
	DeleteFromTestDB(f.T, f.DB, f.Endpoint, f.Channel, f.Offering,
		f.TemplateOffer, f.User, f.Account, f.Product, f.TemplateAccess)
}
