package ept

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/rakyll/statik/fs"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	errValidateMsg = "incorrect validate message"
	eptTempFile    = "/templates/ept.json"
)

var (
	conf struct {
		DB        *data.DBConfig
		Log       *util.LogConfig
		PayServer *pay.Config
		EptTest   *eptTestConfig
	}

	testDB *reform.DB

	timeout = time.Second * 30
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
}

func newFixture(t *testing.T) *testFixture {
	temp := newTemplate(t)
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

func newTemplate(t *testing.T) *data.Template {
	statikFS, err := fs.New()
	if err != nil {
		t.Fatal(err)
	}

	tpl, err := statikFS.Open(eptTempFile)
	if err != nil {
		t.Fatal(err)
	}
	defer tpl.Close()

	schema, err := ioutil.ReadAll(tpl)
	if err != nil {
		t.Fatal(err)
	}

	temp := data.NewTestTemplate(data.TemplateOffer)
	temp.Raw = schema

	return temp
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
	return &eptTestConfig{}
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.PayServer = &pay.Config{}
	conf.EptTest = newEptTestConfig()

	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	testDB = data.NewTestDB(conf.DB, logger)

	defer data.CloseDB(testDB)

	os.Exit(m.Run())
}

func TestValidEndpointMessage(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	s, err := New(testDB, conf.PayServer.Addr)
	if err != nil {
		t.Fatal(err)
	}

	em, err := s.EndpointMessage(fxt.ch.ID, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if em == nil {
		t.Fatal(ErrInvalidFormat)
	}
}

func TestBadProductConfig(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	fxt.product.Config = []byte(`{}`)
	if err := testDB.Update(fxt.product); err != nil {
		t.Fatal(err)
	}

	s, err := New(testDB, conf.PayServer.Addr)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.EndpointMessage(fxt.ch.ID, timeout); err == nil {
		t.Fatal(errValidateMsg)
	}
}

func TestBadProductOfferAccessID(t *testing.T) {
	fxt := newFixture(t)
	defer fxt.clean()

	fxt.product.OfferAccessID = nil
	if err := testDB.Update(fxt.product); err != nil {
		t.Fatal(err)
	}

	s, err := New(testDB, conf.PayServer.Addr)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := s.EndpointMessage(fxt.ch.ID, timeout); err == nil {
		t.Fatal(errValidateMsg)
	}
}
