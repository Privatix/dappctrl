package sess_test

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
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
	client  *rpc.Client
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
	t.Helper()

	fixture := data.NewTestFixture(t, db)
	fixture.Channel.ServiceStatus = data.ServiceActive
	data.SaveToTestDB(t, fixture.DB, fixture.Channel)
	return fixture
}

func newClient(queue job.Queue) *rpc.Client {
	server := rpc.NewServer()
	handler = sess.NewHandler(log.NewMultiLogger(),
		db, newTestCountryConfig(), queue)
	if err := server.RegisterName("sess", handler); err != nil {
		panic(err)
	}
	return rpc.DialInProc(server)
}

func TestMain(m *testing.M) {
	conf.DB = data.NewDBConfig()
	util.ReadTestArgs(&util.TestArgs{Conf: &conf})

	db = data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	client = newClient(job.NewDummyQueueMock())
	defer client.Close()

	os.Exit(m.Run())
}
