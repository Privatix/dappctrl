package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

var offeringRelatedEventsMap = map[common.Hash]bool{
	common.HexToHash(eth.EthOfferingCreated): true,
	common.HexToHash(eth.EthOfferingDeleted): true,
	common.HexToHash(eth.EthOfferingPoppedUp): true,
}

// schedule creates a job for each unprocessed log event in the database.
func (m *Monitor) schedule(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second) // FIXME: hardcoded duration
	defer cancel()

	// TODO: Move this logic into a database view? The query is just supposed to
	// append two boolean columns calculated based on topics: whether the
	// event is for agent, and the same for client.
	//
	// eth_logs.topics is a json array with '0xdeadbeef' encoding of addresses,
	// whereas accounts.eth_addr is a base64 encoding of raw bytes of addresses.
	// The encode-decode-substr is there to convert from one to another.
	// coalesce() converts null into false for the case when topics->>n does not exist.
	topicInAccExpr := `
		coalesce(
			encode(decode(substr(topics->>%d, 3), 'hex'), 'base64')
			in (select eth_addr from accounts where in_use),
			false
		)
	`
	columns := m.db.QualifiedColumns(data.LogEntryTable)
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 1)) // topic[1] (agent) in active accounts
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 2)) // topic[2] (client) in active accounts

	query := fmt.Sprintf(
		"select %s from eth_logs where job IS NULL",
		strings.Join(columns, ","),
	)
	rows, err := m.db.Query(query)
	if err != nil {
		panic(fmt.Errorf("failed to select log entries: %v", err))
	}

	for rows.Next() {
		var e data.LogEntry
		var forAgent, forClient bool
		pointers := append(e.Pointers(), &forAgent, &forClient)
		if err := rows.Scan(pointers...); err != nil {
			panic(fmt.Errorf("failed to scan the selected log entries: %v", err))
		}

		switch {
			case forAgent:
				m.scheduleAgentRelated(&e)
			case forClient:
				m.scheduleClientRelated(&e)
			case isOfferingRelated(&e):
				m.scheduleOfferingRelated(&e)
			default:
				m.ignoreEvent(&e)
		}
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("failed to fetch the next selected log entry: %v", err))
	}
}

func isOfferingRelated(e *data.LogEntry) bool {
	return len(e.Topics) > 0 && offeringRelatedEventsMap[common.HexToHash(e.Topics[0])]
}

var agentEventToJobMap = map[common.Hash]string{
	common.HexToHash(eth.EthDigestChannelCreated): data.JobAgentAfterChannelCreate,
	common.HexToHash(eth.EthDigestChannelToppedUp): data.JobAgentAfterChannelTopUp,
	common.HexToHash(eth.EthChannelCloseRequested): data.JobAgentAfterUncooperativeCloseRequest,
	// common.HexToHash(eth.EthOfferingEndpoint) // FIXME: ignore?
	common.HexToHash(eth.EthCooperativeChannelClose): data.JobAgentAfterCooperativeClose,
	common.HexToHash(eth.EthUncooperativeChannelClose): data.JobAgentAfterUncooperativeClose,
	common.HexToHash(eth.EthOfferingCreated): data.JobAgentAfterOfferingMsgBCPublish,
}

var clientEventToJobMap = map[common.Hash]string{
	common.HexToHash(eth.EthDigestChannelCreated): data.JobClientAfterChannelCreate,
	common.HexToHash(eth.EthDigestChannelToppedUp): data.JobClientAfterChannelTopUp,
	common.HexToHash(eth.EthChannelCloseRequested): data.JobClientAfterUncooperativeCloseRequest,
	// common.HexToHash(eth.EthOfferingEndpoint) // FIXME: ignore?
	common.HexToHash(eth.EthCooperativeChannelClose): data.JobClientAfterCooperativeClose,
	common.HexToHash(eth.EthUncooperativeChannelClose): data.JobClientAfterUncooperativeClose,
}

var offeringEventToJobMap = map[common.Hash]string{
	common.HexToHash(eth.EthOfferingCreated): data.JobClientAfterOfferingMsgBCPublish,
	// common.HexToHash(eth.EthOfferingDeleted) // special case handled by the monitor itself
	common.HexToHash(eth.EthOfferingPoppedUp): data.JobClientAfterOfferingMsgBCPublish,
}

// schedule creates an agent job for a given event.
func (m *Monitor) scheduleAgentRelated(e *data.LogEntry) {
	jobType, found := agentEventToJobMap[common.HexToHash(e.Topics[0])]
	if !found {
		m.ignoreEvent(e)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedType: data.JobChannel,
	}

	switch jobType {
		case data.JobAgentAfterChannelCreate:
			j.RelatedID = util.NewUUID()
		default:
			j.RelatedID = util.NewUUID() // FIXME: query the db
	}

	m.scheduleCommon(e, j)
}

// schedule creates a client job for a given event.
func (m *Monitor) scheduleClientRelated(e *data.LogEntry) {
	jobType, found := clientEventToJobMap[common.HexToHash(e.Topics[0])]
	if !found {
		m.ignoreEvent(e)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedType: data.JobChannel,
	}
	m.scheduleCommon(e, j)
}

// schedule creates a job for a given offering-related event.
func (m *Monitor) scheduleOfferingRelated(e *data.LogEntry) {
	jobType, found := offeringEventToJobMap[common.HexToHash(e.Topics[0])]
	if !found {
		m.ignoreEvent(e)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedType: data.JobOffering,
	}
	m.scheduleCommon(e, j)
}

func (m *Monitor) scheduleCommon(e *data.LogEntry, j *data.Job) {
	j.CreatedBy = data.JobBCMonitor
	j.CreatedAt = time.Now()
	switch err := m.queue.Add(j); err {
		case nil:
			m.updateEventJobID(e, j.ID)
		case job.ErrDuplicatedJob, job.ErrAlreadyProcessing:
			m.ignoreEvent(e)
		default:
			m.incrementEventFailures(e)
	}
}

func (m *Monitor) incrementEventFailures(e *data.LogEntry) {
	e.Failures++
	if err := m.db.UpdateColumns(e, "failures"); err != nil {
		panic(fmt.Errorf("failed to ignore event: %v", err))
	}
}

func (m *Monitor) updateEventJobID(e *data.LogEntry, jobID string) {
	e.JobID = &jobID
	if err := m.db.UpdateColumns(e, "job"); err != nil {
		panic(fmt.Errorf("failed to update job_id of an event: %v", err))
	}
}

func (m *Monitor) ignoreEvent(e *data.LogEntry) {
	m.updateEventJobID(e, "00000000-0000-0000-0000-000000000000")
}

