// +build !nobillingtest

package billing

import (
	"gopkg.in/reform.v1"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"os"
	"time"
)

var (
	testDB  *reform.DB
	testMon *Monitor

	offerSmallMaxUnit  uint64 = 900
	offerUnitPrice     uint64 = 1
	offerBigLag        uint   = 10000
	offerSmallLag      uint   = 1
	sesUnitsUsed       uint64 = 300
	sesSecondsConsumed uint64 = 300
)

const testPassword = "test-password"

type testFixture struct {
	client   *data.User
	agent    *data.Account
	product  *data.Product
	template *data.Template
}

func newFixture(t *testing.T) *testFixture {
	clientAcc := data.NewTestAccount(testPassword)
	client := data.NewTestUser()
	client.PublicKey = clientAcc.PublicKey
	client.EthAddr = clientAcc.EthAddr
	agent := data.NewTestAccount(testPassword)
	product := data.NewTestProduct()
	template := data.NewTestTemplate(data.TemplateOffer)

	data.InsertToTestDB(t, testDB, client, agent, product, template)

	return &testFixture{
		client:   client,
		agent:    agent,
		product:  product,
		template: template,
	}
}

func sesFabric(chanID string, secondsConsumed,
	unitsUsed uint64, adjustTime int64, num int) (sessions []*data.Session) {
	if num <= 0 {
		return sessions
	}

	var curTime time.Time

	if adjustTime != 0 {
		curTime = time.Now().Add(time.Second * time.Duration(adjustTime))
	} else {
		curTime = time.Now()
	}

	for i := 0; i <= num; i++ {
		sessions = append(sessions, &data.Session{
			ID:              util.NewUUID(),
			Channel:         chanID,
			Started:         time.Now(),
			LastUsageTime:   curTime,
			SecondsConsumed: secondsConsumed,
			UnitsUsed:       unitsUsed,
		})
	}
	return sessions
}

func TestMain(m *testing.M) {
	conf := struct {
		DB  *data.DBConfig
		Log *util.LogConfig
	}{
		DB:  data.NewDBConfig(),
		Log: util.NewLogConfig(),
	}
	util.ReadTestConfig(&conf)
	logger := util.NewTestLogger(conf.Log)
	testDB = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(testDB)
	mon, err := NewMonitor(time.Second, testDB, logger)
	if err != nil {
		panic(err)
	}
	if mon == nil {
		panic("Monitor object not created")
	}
	testMon = mon
	os.Exit(m.Run())
}

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
func TestMonitor_VerifySecondsBasedChannels_TotalDeposit(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 1,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 100,
		data.ChannelActive)

	data.InsertToTestDB(t, testDB, offering, channel1, channel2)

	if err := testMon.VerifySecondsBasedChannels(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}

}

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
func TestMonitor_VerifySecondsBasedChannels_MaxUnit(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering.MaxUnit = &offerSmallMaxUnit
	offering.UnitPrice = offerUnitPrice

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 10000,
		data.ChannelActive)
	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 10000,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID,
		sesSecondsConsumed, 0, 0, 3)
	sesChannel2 := sesFabric(channel2.ID,
		sesSecondsConsumed, 0, 0, 2)

	data.InsertToTestDB(t, testDB, offering, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1])

	if err := testMon.VerifySecondsBasedChannels(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}
}

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
func TestMonitor_VerifyUnitsBasedChannels_TotalDeposit(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering.UnitType = data.UnitScalar

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 1,
		data.ChannelActive)
	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 100,
		data.ChannelActive)

	data.InsertToTestDB(t, testDB, offering, channel1, channel2)

	if err := testMon.VerifyUnitsBasedChannels(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}

}

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
func TestMonitor_VerifyUnitsBasedChannels_MaxUnit(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering.MaxUnit = &offerSmallMaxUnit
	offering.UnitPrice = 1
	offering.UnitType = data.UnitScalar

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 10000,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0, 10000,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID, 0,
		sesUnitsUsed, 0, 3)
	sesChannel2 := sesFabric(channel2.ID, 0,
		sesUnitsUsed, 0, 2)

	data.InsertToTestDB(t, testDB, offering, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1])

	if err := testMon.VerifyUnitsBasedChannels(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}
}

// Source conditions:
// There are 2 active channels, that are related to 2 different offerings.
// First offering has relatively big billing lag.
// Seconds one has very small billing lag.
//
// Expected result:
// Channel 1 is not affected.
// Channel 2 is selected for suspending.
func TestMonitor_VerifyBillingLags(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxBillingUnitLag = offerBigLag

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxBillingUnitLag = offerSmallLag

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID, 0, 10000,
		data.ChannelActive)
	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0, 10000,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID, sesUnitsUsed,
		sesUnitsUsed, 0, 3)
	sesChannel2 := sesFabric(channel2.ID, sesUnitsUsed,
		sesUnitsUsed, 0, 2)

	data.InsertToTestDB(t, testDB, offering1, offering2, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1])

	if err := testMon.VerifyBillingLags(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel2.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}
}

// Source conditions:
// There are 2 suspended channels, that are related to 2 different offerings.
// First offering has relatively big billing lag, so on the next check would be interpret as paid.
// Seconds one has very small billing lag, so on the next check would be interpret as not paid.
//
// Expected result:
// Channel 1 is selected for UNsuspending.
// Channel 2 is not affected.
func TestMonitor_VerifySuspendedChannelsAndTryToUnsuspend(t *testing.T) {
	defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxBillingUnitLag = offerBigLag

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxBillingUnitLag = offerSmallLag

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID, 0, 10000,
		data.ChannelActive)
	channel1.ServiceStatus = data.ServiceSuspended

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0, 10000,
		data.ChannelActive)
	channel2.ServiceStatus = data.ServiceSuspended

	sesChannel1 := sesFabric(channel1.ID, sesUnitsUsed,
		sesUnitsUsed, 0, 3)
	sesChannel2 := sesFabric(channel2.ID, sesUnitsUsed,
		sesUnitsUsed, 0, 2)

	data.InsertToTestDB(t, testDB, offering1, offering2, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1])

	if err := testMon.VerifySuspendedChannelsAndTryToUnsuspend(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be unsuspended")
	}
}

// Source conditions:
// There are 2 active channels, that are related to 2 different offerings.
// First offering has several obsolete session records and is inactive.
// Seconds one has no one obsolete session record (but has fresh sessions records as well).
//
// Expected result:
// Channel 1 is selected for suspending.
// Channel 2 is not affected.
/*func TestMonitor_VerifyChannelsForInactivity(t *testing.T) {
	//defer data.CleanTestDB(t, testDB)
	fixture := newFixture(t)

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxUnit = &offerBigMaxUnit
	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxUnit = &offerBigMaxUnit

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID, 0, 10000,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0, 10000,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID, sesUnitsUsed,
		sesUnitsUsed, -100, 2)
	sesChannel2 := sesFabric(channel2.ID, sesUnitsUsed,
		sesUnitsUsed, 0, 2)

	data.InsertToTestDB(t, testDB, offering1, offering2, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1])

	if err := testMon.VerifyChannelsForInactivity(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	t.Log(testMon.testsSelectedChannelsIDs)
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel1.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}
}*/

/*


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
*/
