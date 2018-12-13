package data

import (
	"os"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		DB *DBConfig
	}
	db *reform.DB
)

func TestRecoverServiceStatuses(t *testing.T) {
	fxt := NewTestFixture(t, db)
	defer fxt.Close()

	fxt.Product.IsServer = false
	SaveToTestDB(t, db, fxt.Product)

	for i := 0; i < 2; i++ {
		for _, v := range []string{
			ServiceActivating, ServiceActive, ServiceSuspending} {
			fxt.Channel.ServiceStatus = v
			SaveToTestDB(t, db, fxt.Channel)

			err := Recover(db)
			util.TestExpectResult(t, "Recover", nil, err)

			ReloadFromTestDB(t, db, fxt.Channel)

			same := fxt.Channel.ServiceStatus == v
			susp := fxt.Channel.ServiceStatus == ServiceSuspended
			if (fxt.Product.IsServer && !same) ||
				(!fxt.Product.IsServer && !susp) {
				t.Error("unexpected service status: ",
					fxt.Channel.ServiceStatus)
			}
		}

		fxt.Channel.ServiceStatus = ServiceTerminating
		SaveToTestDB(t, db, fxt.Channel)

		err := Recover(db)
		util.TestExpectResult(t, "Recover", nil, err)

		ReloadFromTestDB(t, db, fxt.Channel)

		same := fxt.Channel.ServiceStatus == ServiceTerminating
		term := fxt.Channel.ServiceStatus == ServiceTerminated
		if (fxt.Product.IsServer && !same) ||
			(!fxt.Product.IsServer && !term) {
			t.Error("unexpected service status: ",
				fxt.Channel.ServiceStatus)
		}

		fxt.Product.IsServer = true
		SaveToTestDB(t, db, fxt.Product)
	}
}

func TestMain(m *testing.M) {
	conf.DB = NewDBConfig()
	util.ReadTestConfig(&conf)

	db = NewTestDB(conf.DB)

	os.Exit(m.Run())
}
