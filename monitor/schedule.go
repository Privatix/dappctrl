package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

const (
	maxRetryKey = "eth.event.maxretry"
)

func mustParseABI(abiJSON string) abi.ABI {
	a, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		panic(err)
	}
	return a
}

var pscABI = mustParseABI(contract.PrivatixServiceContractABI)

var offeringRelatedEventsMap = map[common.Hash]bool{
	common.HexToHash(eth.EthOfferingCreated):  true,
	common.HexToHash(eth.EthOfferingDeleted):  true,
	common.HexToHash(eth.EthOfferingPoppedUp): true,
}

// schedule creates a job for each unprocessed log event in the database.
func (m *Monitor) schedule(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second) // FIXME: hardcoded duration
	defer cancel()

	// TODO: Move this logic into a database view? The query is just supposed to
	// append two boolean columns calculated based on topics: whether the
	// event is for agent, and the same for client.
	//
	// eth_logs.topics is a json array with '0x0..0deadbeef' encoding of addresses,
	// whereas accounts.eth_addr is a base64 encoding of raw bytes of addresses.
	// The encode-decode-substr is there to convert from one to another.
	// coalesce() converts null into false for the case when topics->>n does not exist.
	topicInAccExpr := `
		COALESCE(
			TRANSLATE(
				encode(decode(substr(topics->>%d, 27), 'hex'), 'base64'),
				'+/',
				'-_'
			)
			IN (SELECT eth_addr FROM accounts WHERE in_use),
			FALSE
		)
	`
	columns := m.db.QualifiedColumns(data.EthLogTable)
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 1)) // topic[1] (agent) in active accounts
	columns = append(columns, fmt.Sprintf(topicInAccExpr, 2)) // topic[2] (client) in active accounts

	query := fmt.Sprintf(
		"SELECT %s FROM eth_logs WHERE job IS NULL AND NOT ignore ORDER BY block_number",
		strings.Join(columns, ","),
	)

	var args []interface{}
	maxRetries := m.getUint64Setting(maxRetryKey)
	if maxRetries != 0 {
		query += " AND failures <= $1"
		args = append(args, maxRetries)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		panic(fmt.Errorf("failed to select log entries: %v", err))
	}

	for rows.Next() {
		var el data.EthLog
		var forAgent, forClient bool
		pointers := append(el.Pointers(), &forAgent, &forClient)
		if err := rows.Scan(pointers...); err != nil {
			panic(fmt.Errorf("failed to scan the selected log entries: %v", err))
		}

		eventHash := el.Topics[0]

		var scheduler funcAndType
		found := false
		switch {
		case forAgent:
			scheduler, found = agentSchedulers[eventHash]
		case forClient:
			scheduler, found = clientSchedulers[eventHash]
		case isOfferingRelated(&el):
			scheduler, found = offeringSchedulers[eventHash]
		}

		if !found {
			m.logger.Debug("scheduler not found for event %s", eventHash.Hex())
			m.ignoreEvent(&el)
			continue
		}

		scheduler.f(m, &el, scheduler.t)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("failed to fetch the next selected log entry: %v", err))
	}
}

func isOfferingRelated(el *data.EthLog) bool {
	return len(el.Topics) > 0 && offeringRelatedEventsMap[el.Topics[0]]
}

type scheduleFunc func(*Monitor, *data.EthLog, string)
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

// getOpenBlockNumber extracts the Open_block_number field of a given
// channel-related EthLog. Returns false in case it failed, i.e.
// the event has no such field.
func getOpenBlockNumber(el *data.EthLog) (uint32, bool) {
	bs, err := data.ToBytes(el.Data)
	if err != nil {
		return 0, false
	}

	switch el.Topics[0] {
	case common.HexToHash(eth.EthDigestChannelToppedUp):
		e := new(contract.PrivatixServiceContractLogChannelToppedUp)
		if err := pscABI.Unpack(e, "LogChannelToppedUp", bs); err == nil {
			return e.Open_block_number, true
		} else {
			panic(err)
		}
	case common.HexToHash(eth.EthChannelCloseRequested):
		e := new(contract.PrivatixServiceContractLogChannelCloseRequested)
		if err := pscABI.Unpack(e, "LogChannelCloseRequested", bs); err == nil {
			return e.Open_block_number, true
		} else {
			panic(err)
		}
	case common.HexToHash(eth.EthCooperativeChannelClose):
		e := new(contract.PrivatixServiceContractLogCooperativeChannelClose)
		if err := pscABI.Unpack(e, "LogCooperativeChannelClose", bs); err == nil {
			return e.Open_block_number, true
		} else {
			panic(err)
		}
	case common.HexToHash(eth.EthUncooperativeChannelClose):
		e := new(contract.PrivatixServiceContractLogUnCooperativeChannelClose)
		if err := pscABI.Unpack(e, "LogUnCooperativeChannelClose", bs); err == nil {
			return e.Open_block_number, true
		} else {
			panic(err)
		}
	}

	return 0, false
}

func (m *Monitor) findChannelID(el *data.EthLog) string {
	agentAddress := common.BytesToAddress(el.Topics[1].Bytes())
	clientAddress := common.BytesToAddress(el.Topics[2].Bytes())
	offeringHash := el.Topics[3]

	openBlockNumber, haveOpenBlockNumber := getOpenBlockNumber(el)
	m.logger.Warn("bn = %d, hbn = %t", openBlockNumber, haveOpenBlockNumber)

	var query string
	args := []interface{}{
		data.FromBytes(offeringHash.Bytes()),
		data.FromBytes(agentAddress.Bytes()),
		data.FromBytes(clientAddress.Bytes()),
	}
	if haveOpenBlockNumber {
		query = `
			SELECT c.id
			FROM
				channels AS c,
				offerings AS o
			WHERE
				c.offering = o.id
				AND o.hash = $1
				AND c.agent = $2
				AND c.client = $3
				AND c.block = $4
		`
		args = append(args, openBlockNumber)
	} else {
		query = `
			SELECT c.id
			FROM
				channels AS c,
				offerings AS o,
				eth_txs AS et
			WHERE
				c.offering = o.id
				AND o.hash = $1
				AND c.agent = $2
				AND c.client = $3
				AND et.hash = $4
		`
		args = append(args, el.TxHash)
	}
	row := m.db.QueryRow(query, args...)

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

func (m *Monitor) scheduleAgent_ChannelCreated(el *data.EthLog, jobType string) {
	j := &data.Job{
		Type:        jobType,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleAgentClient_Channel(el *data.EthLog, jobType string) {
	cid := m.findChannelID(el)
	if cid == "" {
		m.logger.Warn("channel for offering %s does not exist", el.Topics[3].Hex())
		m.ignoreEvent(el)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   cid,
		RelatedType: data.JobChannel,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) isOfferingDeleted(offeringHash common.Hash) bool {
	query := `
		SELECT COUNT(*)
		FROM eth_logs
		WHERE
			topics->>0 = $1
			AND topics->>2 = $2
	`
	row := m.db.QueryRow(query, "0x"+eth.EthOfferingDeleted, offeringHash.Hex())

	var count int
	if err := row.Scan(&count); err != nil {
		panic(err)
	}

	return count > 0
}

func (m *Monitor) scheduleClient_OfferingCreated(el *data.EthLog, jobType string) {
	offeringHash := el.Topics[2]
	if m.isOfferingDeleted(offeringHash) {
		m.ignoreEvent(el)
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobOffering,
	}

	m.scheduleCommon(el, j)
}

/* // FIXME: uncomment if monitor should actually delete the offering

// scheduleClient_OfferingDeleted is a special case, which does not
// actually schedule any task, it deletes the offering instead.
func (m *Monitor) scheduleClient_OfferingDeleted(el *data.EthLog, jobType string) {
	offeringHash := common.HexToHash(el.Topics[1])
	tail := "where hash = $1"
	_, err := m.db.DeleteFrom(data.OfferingTable, tail, data.FromBytes(offeringHash.Bytes()))
	if err != nil {
		panic(err)
	}
	m.ignoreEvent(el)
}
*/

func (m *Monitor) scheduleCommon(el *data.EthLog, j *data.Job) {
	j.CreatedBy = data.JobBCMonitor
	j.CreatedAt = time.Now()
	err := m.queue.Add(j)
	switch err {
	case nil:
		m.updateEventJobID(el, j.ID)
	case job.ErrDuplicatedJob, job.ErrAlreadyProcessing:
		m.ignoreEvent(el)
	default:
		m.incrementEventFailures(el)
	}
}

func (m *Monitor) incrementEventFailures(el *data.EthLog) {
	el.Failures++
	if err := m.db.UpdateColumns(el, "failures"); err != nil {
		panic(fmt.Errorf("failed to update failure counter of an event: %v", err))
	}
}

func (m *Monitor) updateEventJobID(el *data.EthLog, jobID string) {
	el.JobID = &jobID
	if err := m.db.UpdateColumns(el, "job"); err != nil {
		panic(fmt.Errorf("failed to update job_id of an event to %s: %v", jobID, err))
	}
}

func (m *Monitor) ignoreEvent(el *data.EthLog) {
	el.Ignore = true
	if err := m.db.UpdateColumns(el, "ignore"); err != nil {
		panic(fmt.Errorf("failed to ignore an event: %v", err))
	}
}
