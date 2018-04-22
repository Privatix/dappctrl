// +build !nobillingtest

package billing

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

type config struct {
	DB  *data.DBConfig
	Log *util.LogConfig
}

var (
	conf    = &config{}
	db      *reform.DB
	monitor *Monitor
	logger  *util.Logger
)

func TestMain(m *testing.M) {
	util.ReadTestConfig(&conf)
	logger = util.NewTestLogger(conf.Log)

	var err error
	db, err = data.NewDB(conf.DB, logger)
	if err != nil {
		panic(err)
	}

	monitor = &Monitor{db, logger, time.Second, make([]string, 0)}

	os.Exit(m.Run())
}

func TestA1(t *testing.T) {
	// Source conditions:
	// There are 2 active SECONDS-based channels.
	// First one has very low "total_deposit", that is less, than offering setup price.
	// Second one has enough "total_deposit", that is greater than offering setup price.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks first rule in HAVING block.

	populateDataAndCallValidation(t, "a1_source_data.sql", monitor.VerifySecondsBasedChannels)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be suspended")
	}
}

func TestA2(t *testing.T) {
	// Source conditions:
	// There are 2 active SECONDS-based channels.
	// First one has 3 sessions records, that used in total more seconds, than is provided by the offering.
	// Second one has 2 sessions records, that used less seconds, than provided by the offering.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks second rule in HAVING block.

	populateDataAndCallValidation(t, "a2_source_data.sql", monitor.VerifySecondsBasedChannels)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be suspended")
	}
}

func TestB1(t *testing.T) {
	// Source conditions:
	// There are 2 active UNITS-based channels.
	// First one has very low "total_deposit", that is less, than offering setup price.
	// Second one has enough "total_deposit", that is greater than offering setup price.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks first rule in HAVING block.

	populateDataAndCallValidation(t, "b1_source_data.sql", monitor.VerifyUnitsBasedChannels)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be suspended")
	}
}

func TestB2(t *testing.T) {
	// Source conditions:
	// There are 2 active UNITS-based channels.
	// First one has 3 sessions records, that used in total more units, than is provided by the offering.
	// Second one has 2 sessions records, that used less seconds, than provided by the offering.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks second rule in HAVING block.

	populateDataAndCallValidation(t, "b2_source_data.sql", monitor.VerifyUnitsBasedChannels)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be suspended")
	}
}

func TestC1(t *testing.T) {
	// Source conditions:
	// There are 2 active channels, that are related to 2 different offerings.
	// First offering has relatively big billing lag.
	// Seconds one has very small billing lag.
	//
	// Expected result:
	// Channel 1 is not affected.
	// Channel 2 is selected for suspending.

	populateDataAndCallValidation(t, "c1_source_data.sql", monitor.VerifyBillingLags)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000002" {
		t.Fatal("Billing ignored channel, that must be suspended")
	}
}

func TestC2(t *testing.T) {
	// Source conditions:
	// There are 2 suspended channels, that are related to 2 different offerings.
	// First offering has relatively big billing lag, so on the next check would be interpret as paid.
	// Seconds one has very small billing lag, so on the next check would be interpret as not paid.
	//
	// Expected result:
	// Channel 1 is selected for UNsuspending.
	// Channel 2 is not affected.

	populateDataAndCallValidation(t, "c2_source_data.sql", monitor.VerifySuspendedChannelsAndTryToUnsuspend)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be unsuspended")
	}
}

func TestD1(t *testing.T) {
	// Source conditions:
	// There are 2 active channels, that are related to 2 different offerings.
	// First offering has several obsolete session records and is inactive.
	// Seconds one has no one obsolete session record (but has fresh sessions records as well).
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.

	populateDataAndCallValidation(t, "d1_source_data.sql", monitor.VerifyChannelsForInactivity)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be unsuspended")
	}
}

func TestE1(t *testing.T) {
	// Source conditions:
	// There is one suspended channel, that was suspended much earlier,
	// than service offering allows, before terminating.
	//
	// Expected result:
	// Channel 1 is selected for terminating.

	populateDataAndCallValidation(t, "e1_source_data.sql", monitor.VerifySuspendedChannelsAndTryToTerminate)
	if len(monitor.testsSelectedChannelsIDs) != 1 ||
		monitor.testsSelectedChannelsIDs[0] != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("Billing ignored channel, that must be unsuspended")
	}
}

func populateDataAndCallValidation(t *testing.T, dataSetFilename string, callback func() error) {
	err := executeSQLFile(path.Join(util.RootPath(), "billing", "tests", dataSetFilename))
	if err != nil {
		t.Fatal("Can't populate source test data. Details: ", err)
	}

	err = callback()
	if err != nil {
		t.Fatal("Can't populate source test data. Details: ", err)
	}
}

func executeSQLFile(filename string) error {
	query, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(query))
	return err
}
