// +build !nobillingtest

package billing

import (
	"testing"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func verifySuspendedChannelsAndTryToTerminate(t *testing.T) {
	if err := mon.VerifySuspendedChannelsAndTryToTerminate(); err != nil {
		t.Fatalf(errDB)
	}
}

func genSuspendedChannelsAndTryToTerminate(t *testing.T) *testFixture {
	fixture := newFixture(t)

	pastTime := time.Now().Add(time.Second * (-100))

	offering := data.NewTestOffering(fixture.agent.EthAddr,
		fixture.product.ID, fixture.template.ID)

	channel := data.NewTestChannel(fixture.agent.EthAddr,
		fixture.client.EthAddr, offering.ID, 0,
		conf.BillingTest.Channel.BigDeposit,
		data.ChannelActive)

	channel.ServiceStatus = data.ServiceSuspended

	channel.ServiceChangedTime = &pastTime

	fixture.addTestObjects([]reform.Record{offering, channel})

	fixture.chs = append(fixture.chs, channel)

	return fixture
}

// Source conditions:
// There is one suspended channel, that was suspended much earlier,
// than service offering allows, before terminating.
//
// Expected result:
// Channel 1 is selected for terminating.
func TestSuspendedChannelsAndTryToTerminate(t *testing.T) {
	fixture := genSuspendedChannelsAndTryToTerminate(t)
	defer fixture.clean()

	fixture.checkJob(t, 0, verifySuspendedChannelsAndTryToTerminate,
		data.JobAgentPreServiceTerminate)

	fixture.checkChanStatus(t, 0, verifySuspendedChannelsAndTryToTerminate,
		data.JobAgentPreServiceTerminate)

	fixture.checkAcc(t, 0, verifySuspendedChannelsAndTryToTerminate,
		data.JobAgentPreServiceTerminate)
}
