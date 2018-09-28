package monitor

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func (m *Monitor) RangeOfInterest(ctx context.Context) (uint64, uint64, error) {
	return m.rangeOfInterest(ctx)
}

func (m *Monitor) AgentQueries(from, to uint64,
	psc, ptc common.Address) ([]ethereum.FilterQuery, error) {
	q, e := m.agentQueries(from, to, psc, ptc)
	return q, e
}

func (m *Monitor) ClientQueries(from, to uint64,
	psc, ptc common.Address) ([]ethereum.FilterQuery, error) {
	q, e := m.clientQueries(from, to, psc, ptc)
	return q, e
}

func (m *Monitor) QueryLogsAndCreateJobs(b queriesBuilderFunc, p JobsProducers) error {
	return m.queryLogsAndCreateJobs(b, p)
}

func (m *Monitor) AgentJobsProducers() JobsProducers {
	return m.agentJobsProducers()
}

func (m *Monitor) ClientJobsProducers() JobsProducers {
	return m.clientJobsProducers()
}
