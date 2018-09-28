package monitor_test

import (
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
)

func assertTopicsEqual(t *testing.T, wanted [][]common.Hash, got [][]common.Hash) {
	if len(got) != len(wanted) {
		t.Fatal("wrong topics")
	}
	for i, iTopics := range wanted {
		if iTopics == nil && got[i] != nil ||
			(iTopics != nil && got[i] == nil) {
			t.Fatal("wrong topic ", i)
		}
		for j, topic := range iTopics {
			if got[i][j] != topic {
				t.Fatal("wrong hash in topic ", i)
			}
		}
	}
}

func assertQueryInRange(t *testing.T, from, to uint64, q ethereum.FilterQuery) {
	v := q.FromBlock.Uint64()
	if from != v {
		t.Fatalf("wanted from block: %v, got: %v", from, v)
	}

	v = q.ToBlock.Uint64()
	if to != v {
		t.Fatalf("wanted to block: %v, got: %v", to, v)
	}
}

func TestQueries(t *testing.T) {
	var from, to uint64 = 1, 100
	pscAddr := common.HexToAddress("0x1")
	ptcAddr := common.HexToAddress("0x2")

	accounts := make([]*data.Account, 5)
	accAdds := make([]common.Hash, 5)
	for i, _ := range accounts {
		accounts[i] = data.NewTestAccount(data.TestPassword)
		accAdds[i] = common.HexToHash(accounts[i].EthAddr)
		data.InsertToTestDB(t, db, accounts[i])
		defer data.DeleteFromTestDB(t, db, accounts[i])
	}

	testClientQueries(t, from, to, pscAddr, ptcAddr, accAdds)
	testAgentQueries(t, from, to, pscAddr, ptcAddr, accAdds)
}

// testClientQueries tests that composed two queries. First for events with user
// account address figuring in topics. Second for events on offerings activity,
// both with user account address in and not in events topics.
func testClientQueries(t *testing.T, from, to uint64, psc, ptc common.Address,
	addressesInDB []common.Hash) {

	queries, _ := mon.ClientQueries(from, to, psc, ptc)

	if len(queries) != 3 {
		t.Fatal("wanted: 3 queries, got: ", len(queries))
	}

	q1 := queries[0]

	assertQueryInRange(t, from, to, q1)

	if len(q1.Addresses) != 2 {
		t.Fatal("must search in psc and ptc")
	}

	assertTopicsEqual(t, [][]common.Hash{
		{
			eth.TokenTransfer,
			eth.ServiceChannelToppedUp,
			eth.ServiceChannelCloseRequested,
		},
		nil,
		addressesInDB,
	}, q1.Topics)

	q2 := queries[1]

	assertQueryInRange(t, from, to, q2)

	if len(q2.Addresses) != 1 || q2.Addresses[0] != ptc {
		t.Fatal("must search in ptc only")
	}

	assertTopicsEqual(t, [][]common.Hash{
		{
			eth.TokenTransfer, eth.TokenApproval,
		},
		addressesInDB,
	}, q2.Topics)

	q3 := queries[2]

	assertQueryInRange(t, from, to, q3)

	if len(q3.Addresses) != 1 || q3.Addresses[0] != psc {
		t.Fatal("must search in psc only")
	}

	assertTopicsEqual(t, [][]common.Hash{
		{
			eth.ServiceChannelCreated, eth.ServiceOfferingCreated,
			eth.ServiceOfferingDeleted, eth.ServiceOfferingPopedUp,
			eth.ServiceCooperativeChannelClose,
			eth.ServiceUnCooperativeChannelClose,
		},
	}, q3.Topics)
}

// testAgentQueries tests that composed one query for events with user account
// address figuring in topics.
func testAgentQueries(t *testing.T, from, to uint64, psc, ptc common.Address,
	addressesInDB []common.Hash) {

	queries, _ := mon.AgentQueries(from, to, psc, ptc)

	if len(queries) != 2 {
		t.Fatal("wanted: 2 query, got: ", len(queries))
	}

	q1 := queries[0]

	assertQueryInRange(t, from, to, q1)

	if len(q1.Addresses) != 2 {
		t.Fatal("must search in psc and ptc")
	}

	assertTopicsEqual(t, [][]common.Hash{
		nil,
		addressesInDB,
	}, q1.Topics)

	q2 := queries[1]

	assertQueryInRange(t, from, to, q2)

	if len(q2.Addresses) != 1 || q2.Addresses[0] != ptc {
		t.Fatal("must search in ptc only")
	}

	assertTopicsEqual(t, [][]common.Hash{
		{eth.TokenTransfer},
		nil,
		addressesInDB,
	}, q2.Topics)
}
