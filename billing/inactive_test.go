// +build !nobillingtest

package billing

import (
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func verifyChannelsForInactivity(t *testing.T) {
	if err := mon.VerifyChannelsForInactivity(); err != nil {
		t.Fatalf(errDB)
	}
}

func genChannelsForInactivity(t *testing.T) *testFixture {
	fixture := newFixture(t)

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering1.MaxInactiveTimeSec =
		&conf.BillingTest.Offer.MaxInactiveTimeSec

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering2.MaxInactiveTimeSec =
		&conf.BillingTest.Offer.MaxInactiveTimeSec

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0,
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

	fixture.chs = append(fixture.chs, channel1, channel2)

	return fixture
}

// Source conditions:
// There are 2 active channels, that are related to 2 different offerings.
// First offering has several obsolete session records and is inactive.
// Seconds one has no one obsolete session record
// (but has fresh sessions records as well).
//
// Expected result:
// Channel 1 is selected for terminating.
// Channel 2 is not affected.
func TestChannelsForInactivity(t *testing.T) {
	fixture := genChannelsForInactivity(t)
	defer fixture.clean()

	fixture.checkJob(t, 0, verifyChannelsForInactivity,
		data.JobAgentPreServiceTerminate)

	fixture.checkChanStatus(t, 0, verifyChannelsForInactivity,
		data.JobAgentPreServiceTerminate)

	fixture.checkAcc(t, 0, verifyChannelsForInactivity,
		data.JobAgentPreServiceTerminate)
}
