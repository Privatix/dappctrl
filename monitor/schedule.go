package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

const (
	maxRetryKey = "eth.event.maxretry"
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
			translate(
				encode(decode(substr(topics->>%d, 3), 'hex'), 'base64'),
				'+/',
				'-_'
			)
			in (select eth_addr from accounts where in_use),
			false
		)
	`
	columns := m.db.QualifiedColumns(data.LogEntryTable)
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 1)) // topic[1] (agent) in active accounts
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 2)) // topic[2] (client) in active accounts

	query := fmt.Sprintf(
		"select %s from eth_logs where job is null and not ignore order by block_number",
		strings.Join(columns, ","),
	)

	var args []interface{}
	maxRetries := m.getUint64Setting(maxRetryKey)
	if maxRetries != 0 {
		query += " and failures <= " + m.db.Placeholder(1)
		args = append(args, maxRetries)
	}

	rows, err := m.db.Query(query, args...)
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

		eventHash := common.HexToHash(e.Topics[0])

		var scheduler funcAndType
		found := false
		switch {
			case forAgent:
				scheduler, found = agentSchedulers[eventHash]
			case forClient:
				scheduler, found = clientSchedulers[eventHash]
			case isOfferingRelated(&e):
				scheduler, found = offeringSchedulers[eventHash]
		}

		if !found {
			m.ignoreEvent(&e)
			continue
		}

		scheduler.f(m, &e, scheduler.t)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("failed to fetch the next selected log entry: %v", err))
	}
}

func isOfferingRelated(e *data.LogEntry) bool {
	return len(e.Topics) > 0 && offeringRelatedEventsMap[common.HexToHash(e.Topics[0])]
}

type scheduleFunc func(*Monitor, *data.LogEntry, string)
type funcAndType struct {
	f scheduleFunc
	t string
}

var agentSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthDigestChannelCreated): {
		(*Monitor).scheduleAgent_ChannelCreated,
		data.JobAgentAfterChannelCreate,
	},
	common.HexToHash(eth.EthDigestChannelToppedUp): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobAgentAfterChannelTopUp,
	},
	common.HexToHash(eth.EthChannelCloseRequested): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobAgentAfterUncooperativeCloseRequest,
	},
	common.HexToHash(eth.EthCooperativeChannelClose): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobAgentAfterCooperativeClose,
	},
	common.HexToHash(eth.EthUncooperativeChannelClose): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobAgentAfterUncooperativeClose,
	},
}

var clientSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthDigestChannelCreated): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobClientAfterChannelCreate,
	},
	common.HexToHash(eth.EthDigestChannelToppedUp): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobClientAfterChannelTopUp,
	},
	common.HexToHash(eth.EthChannelCloseRequested): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobClientAfterUncooperativeCloseRequest,
	},
	common.HexToHash(eth.EthCooperativeChannelClose): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobClientAfterCooperativeClose,
	},
	common.HexToHash(eth.EthUncooperativeChannelClose): {
		(*Monitor).scheduleAgentClient_Channel,
		data.JobClientAfterUncooperativeClose,
	},
}

var offeringSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthOfferingCreated): {
		(*Monitor).scheduleClient_OfferingCreated,
		data.JobClientAfterOfferingMsgBCPublish,
	},
	common.HexToHash(eth.EthOfferingPoppedUp): {
		(*Monitor).scheduleClient_OfferingCreated,
		data.JobClientAfterOfferingMsgBCPublish,
	},
	/* // FIXME: uncomment if monitor should actually delete the offering
	common.HexToHash(eth.EthCOfferingDeleted): {
		(*Monitor).scheduleClient_OfferingDeleted,
		"",
	},
	*/
}

func (m *Monitor) findChannelID(e *data.LogEntry) string {
	agentAddress := common.HexToAddress(e.Topics[1])
	clientAddress := common.HexToAddress(e.Topics[2])
	offeringHash := common.HexToHash(e.Topics[3])
	query := fmt.Sprintf(`
		select channels.id
		from channels, offerings
		where
			channels.offering = offerings.id
			and offerings.hash = %s
			and channels.agent = %s
			and channels.client = %s
	`, m.db.Placeholder(1), m.db.Placeholder(2), m.db.Placeholder(3))

	row := m.db.QueryRow(
		query,
		data.FromBytes(offeringHash.Bytes()),
		data.FromBytes(agentAddress.Bytes()),
		data.FromBytes(clientAddress.Bytes()),
	)

	var id string
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return ""
	}
	if err != nil {
		panic(err)
	}

	return id
}

func (m *Monitor) scheduleAgent_ChannelCreated(e *data.LogEntry, jobType string) {
	j := &data.Job{
		Type:        jobType,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
	}

	m.scheduleCommon(e, j)
}

func (m *Monitor) scheduleAgentClient_Channel(e *data.LogEntry, jobType string) {
	cid := m.findChannelID(e)
	if cid == "" {
		m.logger.Warn("channel for offering %s does not exist", e.Topics[3])
		m.ignoreEvent(e)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   cid,
		RelatedType: data.JobChannel,
	}

	m.scheduleCommon(e, j)
}

func (m *Monitor) isOfferingDeleted(offeringHash common.Hash) bool {
	query := fmt.Sprintf(`
		select count(*)
		from eth_logs
		where
			topics->>0 = %s
			and topics->>2 = %s
	`, m.db.Placeholder(1), m.db.Placeholder(2))
	row := m.db.QueryRow(query, "0x" + eth.EthOfferingDeleted, offeringHash.Hex())

	var count int
	if err := row.Scan(&count); err != nil {
		panic(err)
	}

	return count > 0
}

func (m *Monitor) scheduleClient_OfferingCreated(e *data.LogEntry, jobType string) {
	offeringHash := common.HexToHash(e.Topics[2])
	if m.isOfferingDeleted(offeringHash) {
		m.ignoreEvent(e)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobOffering,
	}

	m.scheduleCommon(e, j)
}

/* // FIXME: uncomment if monitor should actually delete the offering

// scheduleClient_OfferingDeleted is a special case, which does not
// actually schedule any task, it deletes the offering instead.
func (m *Monitor) scheduleClient_OfferingDeleted(e *data.LogEntry, jobType string) {
	offeringHash := common.HexToHash(e.Topics[1])
	tail := fmt.Sprintf(
		"where hash = %s",
		m.db.Placeholder(1),
	)
	_, err := m.db.DeleteFrom(data.OfferingTable, tail, data.FromBytes(offeringHash.Bytes()))
	if err != nil {
		panic(err)
	}
	m.ignoreEvent(e)
}
*/

func (m *Monitor) scheduleCommon(e *data.LogEntry, j *data.Job) {
	j.CreatedBy = data.JobBCMonitor
	j.CreatedAt = time.Now()
	err := m.queue.Add(j)
	switch err {
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
		panic(fmt.Errorf("failed to update failure counter of an event: %v", err))
	}
}

func (m *Monitor) updateEventJobID(e *data.LogEntry, jobID string) {
	e.JobID = &jobID
	if err := m.db.UpdateColumns(e, "job"); err != nil {
		panic(fmt.Errorf("failed to update job_id of an event to %s: %v", jobID, err))
	}
}

func (m *Monitor) ignoreEvent(e *data.LogEntry) {
	e.Ignore = true
	if err := m.db.UpdateColumns(e, "ignore"); err != nil {
		panic(fmt.Errorf("failed to ignore an event: %v", err))
	}
}

