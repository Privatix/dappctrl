package sess_test

import (
	"os"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

var (
	conf struct {
		DB *data.DBConfig
	}

	db      *reform.DB
	handler *sess.Handler
)

func newTestCountryConfig() *country.Config {
	const countryField = "testCountry"

	cs := country.NewServerMock(countryField, "YY")
	defer cs.Close()

	conf := country.NewConfig()
	conf.Field = countryField
	conf.URLTemplate = cs.Server.URL

	return conf
}

func newTestFixture(t *testing.T) *data.TestFixture {
	fixture := data.NewTestFixture(t, db)
	fixture.Channel.ServiceStatus = data.ServiceActive
	if err := db.Update(fixture.Channel); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
<<<<<<< HEAD
	args := &util.TestArgs{
		Conf: &conf,
	}
	util.ReadTestArgs(args)
=======
	util.ReadTestArgs(&util.TestArgs{Conf: &conf})
>>>>>>> Implement ConnChange subscriptions

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	handler = sess.NewHandler(log.NewMultiLogger(),
		db, newTestCountryConfig(), job.NewDummyQueueMock())

	os.Exit(m.Run())
}
