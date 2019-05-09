package ept

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB        *data.DBConfig
		Log       *log.WriterConfig
		PayServer *pay.Config
		EptTest   *eptTestConfig
		EptMsg    *Config
	}

	logger log.Logger

	testDB *reform.DB
)

type testFixture struct {
	t        *testing.T
	product  *data.Product
	template *data.Template
	ch       *data.Channel
	offer    *data.Offering
}

type eptTestConfig struct {
	ServerConfig map[string]string
	Host         string
}

func newFixture(t *testing.T) *testFixture {
	temp := data.NewTestTemplate(data.TemplateAccess)
	prod := newProduct(t, temp.ID)
	offer := newOffer(prod.ID, temp.ID)
	ch := newChan(offer.ID)
	data.InsertToTestDB(t, testDB, temp, prod, offer, ch)

	return &testFixture{
		t:        t,
		template: temp,
		product:  prod,
		offer:    offer,
		ch:       ch,
	}
}

func (f *testFixture) clean() {
	records := append([]reform.Record{}, f.ch, f.offer,
		f.product, f.template)
	for _, v := range records {
		if err := testDB.Delete(v); err != nil {
			f.t.Fatalf("failed to delete %T: %s", v, err)
		}
	}
}

func newProduct(t *testing.T, tempID string) *data.Product {
	prod := data.NewTestProduct()
	prod.OfferAccessID = &tempID

	conf, err := json.Marshal(conf.EptTest.ServerConfig)
	if err != nil {
		t.Fatal(err)
	}

	prod.Config = conf

	return prod
}

func newOffer(prod, tpl string) *data.Offering {
	return data.NewTestOffering("", prod, tpl)
}

func newChan(offer string) *data.Channel {
	return data.NewTestChannel("", "", offer, 100, 100, data.ChannelActive)
}

func newEptTestConfig() *eptTestConfig {
	return &eptTestConfig{
		ServerConfig: make(map[string]string),
		Host:         "localhost:80",
	}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.PayServer = &pay.Config{}
	conf.EptTest = newEptTestConfig()
	conf.EptMsg = NewConfig()
	conf.Log = log.NewWriterConfig()
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)

	testDB = data.NewTestDB(conf.DB)
	defer data.CloseDB(testDB)

	var err error
	logger, err = log.NewTestLogger(conf.Log, args.Verbose)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestValidEndpointMessage(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	fxt.product.ServiceEndpointAddress = &strings.Split(
		conf.EptTest.Host, ":")[0]

	if err := testDB.Update(fxt.product); err != nil {
		t.Fatal(err)
	}

	s, err := New(testDB, logger, conf.PayServer.Addr, conf.EptMsg.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.EndpointMessage(fxt.ch.ID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadProductConfig(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	fxt.product.Config = []byte(`{}`)
	if err := testDB.Update(fxt.product); err != nil {
		t.Fatal(err)
	}

	s, err := New(testDB, logger, conf.PayServer.Addr, conf.EptMsg.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.EndpointMessage(fxt.ch.ID); err == nil {
		t.Fatal(err)
	}
}

func TestBadProductOfferAccessID(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	fxt.product.OfferAccessID = nil
	if err := testDB.Update(fxt.product); err != nil {
		t.Fatal(err)
	}

	s, err := New(testDB, logger, conf.PayServer.Addr, conf.EptMsg.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.EndpointMessage(fxt.ch.ID); err == nil {
		t.Fatal(err)
	}
}
