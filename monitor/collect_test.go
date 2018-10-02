// +build !nomonitortest

package monitor

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

func TestMonitorLogCollect(t *testing.T) {
	defer cleanDB(t)

	mon, _, client := newTestObjects(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := newMockTicker()
	signals := mon.start(ctx, ticker.C, nil)

	_, agentAddress := insertNewAccount(t, db, agentPass)
	_, clientAddress := insertNewAccount(t, db, clientPass)

	eventAboutChannel := common.HexToHash(eth.EthDigestChannelCreated)
	eventAboutOffering := common.HexToHash(eth.EthOfferingCreated)
	eventAboutToken := common.HexToHash(eth.EthTokenApproval)

	var block uint64 = 10

	dataMap := make(map[string]bool)

	type logToInject struct {
		event  common.Hash
		agent  common.Address
		client common.Address
	}
	logsToInject := []logToInject{
		{eventAboutOffering, someAddress, someAddress}, // 1 match all offerings
		{someHash, someAddress, someAddress},           // 0 no match
		{someHash, agentAddress, someAddress},          // 1 match agent
		{someHash, someAddress, clientAddress},         // 0 match client, but not a client event
		// ----- 6 confirmations
		{eventAboutOffering, someAddress, someAddress},  // 1 match all offerings
		{eventAboutChannel, someAddress, someAddress},   // 1 match (for current_supply updates)
		{eventAboutToken, someAddress, someAddress},     // 0 no match
		{eventAboutChannel, agentAddress, someAddress},  // 1 match agent
		{eventAboutChannel, someAddress, clientAddress}, // 1 match client
		// ----- 2 confirmations
		{eventAboutOffering, agentAddress, someAddress}, // 1 match agent
		{eventAboutOffering, someAddress, someAddress},  // 1 match all offerings
		// ----- 0 confirmations
	}
	for _, contractAddr := range []common.Address{someAddress, pscAddr} {
		for _, log := range logsToInject {
			d := genRandData(32 * 5)
			dataMap[data.FromBytes(d)] = true
			var txHash common.Hash
			copy(txHash[:], genRandData(32))
			client.injectEvent(&ethtypes.Log{
				TxHash:      txHash,
				Address:     contractAddr,
				BlockNumber: block,
				Topics: []common.Hash{
					log.event,
					log.agent.Hash(),
					log.client.Hash(),
				},
				Data: d,
			})
			block++
		}
	}

	cases := []struct {
		confirmations uint64
		freshnum      uint64
		lognum        int
	}{
		{6, 2, 2}, // freshnum = 2: will skip the first offering event
		{2, 0, 5}, // freshnum = 0: will include the second offering event and channel event
		{0, 2, 7},
	}

	var logs []*data.EthLog
	for _, c := range cases {
		setUint64Setting(t, db, data.SettingMinConfirmations, c.confirmations)
		setUint64Setting(t, db, data.SettingFreshBlocks, c.freshnum)

		wg := waitSignal(signals.collect)
		ticker.tick()

		wg.Wait()

		name := fmt.Sprintf("with %d confirmations and %d freshnum",
			c.confirmations, c.freshnum)
		logs = expectLogs(t, c.lognum, name, "")
	}

	for _, e := range logs {
		if !dataMap[e.Data] {
			t.Fatal("wrong data saved in a log entry")
		}
		delete(dataMap, e.Data)
	}
}
