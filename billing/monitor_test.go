// +build !nobillingtest

package billing

import (
	"os"
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		DB          *data.DBConfig
		Log         *util.LogConfig
		BillingTest *billingTestConfig
	}
	testDB  *reform.DB
	testMon *Monitor
)

const testPassword = "test-password"

type billingTestConfig struct {
	Offer struct {
		MaxUnit            uint64
		MaxInactiveTimeSec uint64
		UnitPrice          uint64
		BigLag             uint
		SmallLag           uint
	}
	Session struct {
		UnitsUsed            uint64
		EmptyUnitsUsed       uint64
		SecondsConsumed      uint64
		EmptySecondsConsumed uint64
	}
	Channel struct {
		SmallDeposit   uint64
		MidDeposit     uint64
		BigDeposit     uint64
		EmptyBalance   uint64
		EmptyUnitsUsed uint64
	}
}

type testFixture struct {
	t        *testing.T
	client   *data.User
	agent    *data.Account
	product  *data.Product
	template *data.Template
	testObjs []reform.Record
}

func newBillingTestConfig() *billingTestConfig {
	return &billingTestConfig{}
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
		t:        t,
		client:   client,
		agent:    agent,
		product:  product,
		template: template,
	}
}

func (f *testFixture) addTestObjects(testObjs []reform.Record) {
	data.SaveToTestDB(f.t, testDB, testObjs...)
	f.testObjs = testObjs
}

func (f *testFixture) clean() {
	records := append([]reform.Record{}, f.client, f.agent,
		f.product, f.template)
	records = append(records, f.testObjs...)
	reverse(records)
	for _, v := range records {
		if err := testDB.Delete(v); err != nil {
			f.t.Fatalf("failed to delete %T: %s", v, err)
		}
	}
}

func reverse(rs []reform.Record) {
	last := len(rs) - 1
	for i := 0; i < len(rs)/2; i++ {
		rs[i], rs[last-i] = rs[last-i], rs[i]
	}
}

func sesFabric(chanID string, secondsConsumed,
	unitsUsed uint64, adjustTime int64, num int) (
	sessions []*data.Session) {
	if num <= 0 {
		return sessions
	}
	for i := 0; i <= num; i++ {
		curTime := time.Now()
		if adjustTime != 0 {
			curTime = curTime.Add(
				time.Second * time.Duration(adjustTime))
		}
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
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.BillingTest = newBillingTestConfig()
	util.ReadTestConfig(&conf)

	logger := util.NewTestLogger(conf.Log)
	testDB = data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(testDB)

	if _, err := NewMonitor(time.Second, nil, logger); err == nil {
		panic("Monitor object with empty database is created")
	}

	mon, err := NewMonitor(time.Second, testDB, logger)
	if err != nil {
		panic(err)
	}
	testMon = mon
	os.Exit(m.Run())
}

func TestMonitor_VerifySecondsBasedChannels(t *testing.T) {

	// Source conditions:
	// There are 2 active SECONDS-based channels.
	// First one has very low "total_deposit", that is less,
	// than offering setup price.
	// Second one has enough "total_deposit", that is greater
	// than offering setup price.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks first rule in HAVING block.
	t.Run("LowTotalDeposit", func(t *testing.T) {
		fixture := newFixture(t)
		defer fixture.clean()

		offering := data.NewTestOffering(fixture.agent.EthAddr,
			fixture.product.ID, fixture.template.ID)
		channel1 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.SmallDeposit,
			data.ChannelActive)
		channel2 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.MidDeposit,
			data.ChannelActive)

		fixture.addTestObjects([]reform.Record{
			offering, channel1, channel2})

		if err := testMon.VerifySecondsBasedChannels(); err != nil {
			t.Fatalf("Failed to read channel information" +
				" from the database")
		}
		if len(testMon.testsSelectedChannelsIDs) != 1 ||
			testMon.testsSelectedChannelsIDs[0] != channel1.ID {
			t.Fatal("Billing ignored channel," +
				" that must be suspended")
		}
	})

	// Source conditions:
	// There are 2 active SECONDS-based channels.
	// First one has 3 sessions records, that used in total more seconds,
	// than is provided by the offering.
	// Second one has 2 sessions records, that used less seconds,
	// than provided by the offering.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks second rule in HAVING block.
	t.Run("UnitLimitExceeded", func(t *testing.T) {
		fixture := newFixture(t)
		defer fixture.clean()

		offering := data.NewTestOffering(fixture.agent.EthAddr,
			fixture.product.ID, fixture.template.ID)
		offering.MaxUnit = &conf.BillingTest.Offer.MaxUnit
		offering.UnitPrice = conf.BillingTest.Offer.UnitPrice

		channel1 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.BigDeposit,
			data.ChannelActive)
		channel2 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.BigDeposit,
			data.ChannelActive)
		sesChannel1 := sesFabric(channel1.ID,
			conf.BillingTest.Session.SecondsConsumed,
			conf.BillingTest.Session.EmptyUnitsUsed,
			0, 3)
		sesChannel2 := sesFabric(channel2.ID,
			conf.BillingTest.Session.SecondsConsumed,
			conf.BillingTest.Session.EmptyUnitsUsed,
			0, 2)

		fixture.addTestObjects([]reform.Record{
			offering, channel1, channel2,
			sesChannel1[0], sesChannel1[1], sesChannel1[2],
			sesChannel2[0], sesChannel2[1]},
		)

		if err := testMon.VerifySecondsBasedChannels(); err != nil {
			t.Fatalf("Failed to read channel information" +
				" from the database")
		}
		if len(testMon.testsSelectedChannelsIDs) != 1 ||
			testMon.testsSelectedChannelsIDs[0] != channel1.ID {
			t.Fatal("Billing ignored channel," +
				" that must be suspended")
		}
	})
}

func TestMonitor_VerifyUnitsBasedChannels(t *testing.T) {

	// Source conditions:
	// There are 2 active UNITS-based channels.
	// First one has very low "total_deposit", that is less,
	// than offering setup price.
	// Second one has enough "total_deposit",
	// that is greater than offering setup price.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks first rule in HAVING block.
	t.Run("LowTotalDeposit", func(t *testing.T) {
		fixture := newFixture(t)
		defer fixture.clean()

		offering := data.NewTestOffering(fixture.agent.EthAddr,
			fixture.product.ID, fixture.template.ID)
		offering.UnitType = data.UnitScalar

		channel1 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.SmallDeposit,
			data.ChannelActive)
		channel2 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.MidDeposit,
			data.ChannelActive)

		fixture.addTestObjects([]reform.Record{
			offering, channel1, channel2})

		if err := testMon.VerifyUnitsBasedChannels(); err != nil {
			t.Fatalf("Failed to read channel information" +
				" from the database")
		}
		if len(testMon.testsSelectedChannelsIDs) != 1 ||
			testMon.testsSelectedChannelsIDs[0] != channel1.ID {
			t.Fatal("Billing ignored channel," +
				" that must be suspended")
		}
	})

	// Source conditions:
	// There are 2 active UNITS-based channels.
	// First one has 3 sessions records, that used in total more units,
	// than is provided by the offering.
	// Second one has 2 sessions records, that used less seconds,
	// than provided by the offering.
	//
	// Expected result:
	// Channel 1 is selected for suspending.
	// Channel 2 is not affected.
	//
	// Description: this test checks second rule in HAVING block.
	t.Run("UnitLimitExceeded", func(t *testing.T) {
		fixture := newFixture(t)
		defer fixture.clean()

		offering := data.NewTestOffering(fixture.agent.EthAddr,
			fixture.product.ID, fixture.template.ID)
		offering.MaxUnit = &conf.BillingTest.Offer.MaxUnit
		offering.UnitPrice = conf.BillingTest.Offer.UnitPrice
		offering.UnitType = data.UnitScalar

		channel1 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.BigDeposit,
			data.ChannelActive)

		channel2 := data.NewTestChannel(fixture.agent.EthAddr,
			fixture.client.EthAddr, offering.ID,
			conf.BillingTest.Channel.EmptyBalance,
			conf.BillingTest.Channel.BigDeposit,
			data.ChannelActive)

		sesChannel1 := sesFabric(channel1.ID,
			conf.BillingTest.Session.EmptySecondsConsumed,
			conf.BillingTest.Session.UnitsUsed,
			0, 3)
		sesChannel2 := sesFabric(channel2.ID,
			conf.BillingTest.Session.EmptySecondsConsumed,
			conf.BillingTest.Session.UnitsUsed,
			0, 2)

		fixture.addTestObjects([]reform.Record{
			offering, channel1, channel2,
			sesChannel1[0], sesChannel1[1], sesChannel1[2],
			sesChannel2[0], sesChannel2[1]},
		)

		if err := testMon.VerifyUnitsBasedChannels(); err != nil {
			t.Fatalf("Failed to read channel information" +
				" from the database")
		}
		if len(testMon.testsSelectedChannelsIDs) != 1 ||
			testMon.testsSelectedChannelsIDs[0] != channel1.ID {
			t.Fatal("Billing ignored channel," +
				" that must be suspended")
		}
	})
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
	fixture := newFixture(t)
	defer fixture.clean()

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxBillingUnitLag = conf.BillingTest.Offer.BigLag

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxBillingUnitLag = conf.BillingTest.Offer.SmallLag

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, 0, 3)
	sesChannel2 := sesFabric(channel2.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, 0, 2)

	fixture.addTestObjects([]reform.Record{offering1, offering2,
		channel1, channel2, sesChannel1[0],
		sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1]},
	)

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
// First offering has relatively big billing lag, so on the next check
// would be interpret as paid.
// Seconds one has very small billing lag, so on the next check
// would be interpret as not paid.
//
// Expected result:
// Channel 1 is selected for UNsuspending.
// Channel 2 is not affected.
func TestMonitor_VerifySuspendedChannelsAndTryToUnsuspend(t *testing.T) {
	fixture := newFixture(t)
	defer fixture.clean()

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxBillingUnitLag = conf.BillingTest.Offer.BigLag

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxBillingUnitLag = conf.BillingTest.Offer.SmallLag

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel1.ServiceStatus = data.ServiceSuspended

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel2.ServiceStatus = data.ServiceSuspended

	sesChannel1 := sesFabric(channel1.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, 0, 3)
	sesChannel2 := sesFabric(channel2.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, 0, 2)

	fixture.addTestObjects([]reform.Record{offering1, offering2,
		channel1, channel2, sesChannel1[0],
		sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1]},
	)

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
// Seconds one has no one obsolete session record
// (but has fresh sessions records as well).
//
// Expected result:
// Channel 1 is selected for suspending.
// Channel 2 is not affected.
func TestMonitor_VerifyChannelsForInactivity(t *testing.T) {
	fixture := newFixture(t)
	defer fixture.clean()

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering1.MaxInactiveTimeSec =
		&conf.BillingTest.Offer.MaxInactiveTimeSec

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)
	offering2.MaxInactiveTimeSec =
		&conf.BillingTest.Offer.MaxInactiveTimeSec

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, -100, 2)

	sesChannel2 := sesFabric(channel2.ID,
		conf.BillingTest.Session.SecondsConsumed,
		conf.BillingTest.Session.UnitsUsed, 0, 2)

	fixture.addTestObjects([]reform.Record{offering1, offering2,
		channel1, channel2, sesChannel1[0], sesChannel1[1],
		sesChannel2[0], sesChannel2[1]})

	if err := testMon.VerifyChannelsForInactivity(); err != nil {
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
// There is one suspended channel, that was suspended much earlier,
// than service offering allows, before terminating.
//
// Expected result:
// Channel 1 is selected for terminating.
func TestMonitor_VerifySuspendedChannelsAndTryToTerminate(t *testing.T) {
	fixture := newFixture(t)
	defer fixture.clean()
	pastTime := time.Now().Add(time.Second * (-100))

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	channel := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID,
		conf.BillingTest.Channel.EmptyBalance,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel.ServiceStatus = data.ServiceSuspended
	channel.ServiceChangedTime = &pastTime

	fixture.addTestObjects([]reform.Record{offering, channel})

	if err := testMon.VerifySuspendedChannelsAndTryToTerminate(); err != nil {
		t.Fatalf("Failed to read channel information" +
			" from the database")
	}
	if len(testMon.testsSelectedChannelsIDs) != 1 ||
		testMon.testsSelectedChannelsIDs[0] != channel.ID {
		t.Fatal("Billing ignored channel," +
			" that must be suspended")
	}
}
