// +build !nobillingtest

package billing

import (
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func verifyBillingLags(t *testing.T) {
	if err := mon.VerifyBillingLags(); err != nil {
		t.Fatalf(errDB)
	}
}

func genBillingLags(t *testing.T) *testFixture {
	fixture := newFixture(t)

	offering1 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering1.MaxBillingUnitLag = conf.BillingTest.Offer.BigLag

	offering2 := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	offering2.MaxBillingUnitLag = conf.BillingTest.Offer.SmallLag

	channel1 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering1.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel1.ServiceStatus = data.ServiceActive

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)
	channel2.ServiceStatus = data.ServiceActive

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

	fixture.chs = append(fixture.chs, channel1, channel2)

	return fixture
}

// Source conditions:
// There are 2 active channels, that are related to 2 different offerings.
// First offering has relatively big billing lag.
// Seconds one has very small billing lag.
//
// Expected result:
// Channel 1 is not affected.
// Channel 2 is selected for suspending.
func TestBillingLags(t *testing.T) {
	fixture := genBillingLags(t)
	defer fixture.clean()

	fixture.checkJob(t, 1, verifyBillingLags,
		data.JobAgentPreServiceSuspend)

	fixture.checkChanStatus(t, 1, verifyBillingLags,
		data.JobAgentPreServiceSuspend)

	fixture.checkAcc(t, 1, verifyBillingLags,
		data.JobAgentPreServiceSuspend)
}
