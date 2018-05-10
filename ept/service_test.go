package ept

import (
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/util"
)

const errEqual = "The endpoint message template ID in the database" +
	" is not equal to the value obtained"

var (
	conf struct {
		DB        *data.DBConfig
		Log       *util.LogConfig
		PayServer *pay.Config
	}

	testDB *reform.DB

	timeout = time.Second * 30
)

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.PayServer = &pay.Config{}

	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)

	testDB = data.NewTestDB(conf.DB, logger)

	defer data.CloseDB(testDB)

	os.Exit(m.Run())

}

func TestNewEndpointMessageTemplate(t *testing.T) {
	fxt := data.NewTestFixture(t, testDB)

	defer fxt.Close()

	uuid := fxt.Template.ID

	fxt.Product.OfferAccessID = &uuid

	if err := testDB.Save(fxt.Product); err != nil {
		t.Fatal(err)
	}

	s := New(testDB, conf.PayServer)

	p, err := s.EndpointMessageTemplate(fxt.Channel.ID, timeout)
	if err != nil {
		t.Fatal(err)
	}

	var template data.EndpointMessageTemplate

	if err := s.db.FindByPrimaryKeyTo(&template, p); err != nil {
		t.Fatal(err)
	}

	if template.ID != p {
		t.Fatal(errEqual)
	}

	// necessary measure associated with the database schema
	fxt.Product.OfferAccessID = nil

	if err := testDB.Save(fxt.Product); err != nil {
		t.Fatal(err)
	}
}
