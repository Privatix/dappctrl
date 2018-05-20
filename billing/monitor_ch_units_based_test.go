// +build !nobillingtest

package billing

import (
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func verifyUnitsBasedChannels(t *testing.T) {
	if err := mon.VerifyUnitsBasedChannels(); err != nil {
		t.Fatalf(errDB)
	}
}

func genUnitsBasedChannelsLowTotalDeposit(t *testing.T) *testFixture {
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering.UnitType = data.UnitScalar

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0,
		conf.BillingTest.Channel.SmallDeposit,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0,
		conf.BillingTest.Channel.MidDeposit,
		data.ChannelActive)

	fixture.addTestObjects([]reform.Record{
		offering, channel1, channel2})

	fixture.chs = append(fixture.chs, channel1, channel2)

	return fixture
}

func genUnitsBasedChannelsUnitLimitExceeded(t *testing.T) *testFixture {
	fixture := newFixture(t)

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering.MaxUnit = &conf.BillingTest.Offer.MaxUnit

	offering.UnitPrice = conf.BillingTest.Offer.UnitPrice

	offering.UnitType = data.UnitScalar

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	sesChannel1 := sesFabric(channel1.ID, 0,
		conf.BillingTest.Session.UnitsUsed,
		0, 3)

	sesChannel2 := sesFabric(channel2.ID, 0,
		conf.BillingTest.Session.UnitsUsed,
		0, 2)

	fixture.addTestObjects([]reform.Record{
		offering, channel1, channel2,
		sesChannel1[0], sesChannel1[1], sesChannel1[2],
		sesChannel2[0], sesChannel2[1]},
	)

	fixture.chs = append(fixture.chs, channel1, channel2)

	return fixture
}

// Source conditions:
// There are 2 active UNITS-based channels.
// First one has very low "total_deposit", that is less,
// than offering setup price.
// Second one has enough "total_deposit",
// that is greater than offering setup price.
//
// Expected result:
// Channel 1 is selected for terminating.
// Channel 2 is not affected.
//
// Description: this test checks first rule in HAVING block.
func TestUnitsBasedChannelsLowTotalDeposit(t *testing.T) {
	fixture := genUnitsBasedChannelsLowTotalDeposit(t)
	defer fixture.clean()

	fixture.checkJob(t, 0, verifyUnitsBasedChannels,
		data.JobAgentPreServiceTerminate)
}

// Source conditions:
// There are 2 active UNITS-based channels.
// First one has 3 sessions records, that used in total more units,
// than is provided by the offering.
// Second one has 2 sessions records, that used less seconds,
// than provided by the offering.
//
// Expected result:
// Channel 1 is selected for terminating.
// Channel 2 is not affected.
//
// Description: this test checks second rule in HAVING block.
func TestUnitsBasedChannelsUnitLimitExceeded(t *testing.T) {
	fixture := genUnitsBasedChannelsUnitLimitExceeded(t)
	defer fixture.clean()

	fixture.checkJob(t, 0, verifyUnitsBasedChannels,
		data.JobAgentPreServiceTerminate)

	fixture.checkChanStatus(t, 0, verifyUnitsBasedChannels,
		data.JobAgentPreServiceTerminate)

	fixture.checkAcc(t, 0, verifyUnitsBasedChannels,
		data.JobAgentPreServiceTerminate)
}
