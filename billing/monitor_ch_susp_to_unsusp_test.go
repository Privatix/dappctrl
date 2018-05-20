// +build !nobillingtest

package billing

import (
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func verifySuspendedChannelsAndTryToUnsuspend(t *testing.T) {
	if err := mon.VerifySuspendedChannelsAndTryToUnsuspend(); err != nil {
		t.Fatalf(errDB)
	}
}

func genSuspendedChannelsToUnsuspend(t *testing.T) *testFixture {
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

	channel1.ServiceStatus = data.ServiceSuspended

	channel2 := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering2.ID, 0,
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

	fixture.chs = append(fixture.chs, channel1, channel2)

	return fixture
}

// Source conditions:
// There are 2 suspended channels, that are related to 2 different offerings.
// First offering has relatively big billing lag, so on the next check
// would be interpret as paid.
// Seconds one has very small billing lag, so on the next check
// would be interpret as not paid.
//
// Expected result:
// Channel 1 is selected for Unsuspending (Activate).
// Channel 2 is not affected.
func TestSuspendedChannelsToUnsuspend(t *testing.T) {
	fixture := genSuspendedChannelsToUnsuspend(t)
	defer fixture.clean()

	fixture.checkJob(t, 0, verifySuspendedChannelsAndTryToUnsuspend,
		data.JobAgentPreServiceUnsuspend)

	fixture.checkChanStatus(t, 0, verifySuspendedChannelsAndTryToUnsuspend,
		data.JobAgentPreServiceUnsuspend)

	fixture.checkAcc(t, 0, verifySuspendedChannelsAndTryToUnsuspend,
		data.JobAgentPreServiceUnsuspend)
}
