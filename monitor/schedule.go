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

	logChannelToppedUp           = "LogChannelToppedUp"
	logChannelCloseRequested     = "LogChannelCloseRequested"
	logCooperativeChannelClose   = "LogCooperativeChannelClose"
	logUnCooperativeChannelClose = "LogUnCooperativeChannelClose"
)

const (
	topic1 = iota + 1
	topic2
	topic3
)

var offeringRelatedEventsMap = map[common.Hash]bool{
	common.HexToHash(eth.EthOfferingCreated):  true,
	common.HexToHash(eth.EthOfferingDeleted):  true,
	common.HexToHash(eth.EthOfferingPoppedUp): true,
}

// schedule creates a job for each unprocessed log event in the database.
func (m *Monitor) schedule(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(m.cfg.ScheduleTimeout)*time.Second)
	defer cancel()

	topicInAccExpr := `COALESCE(substr(topics->>%d, 27) IN (SELECT eth_addr FROM accounts WHERE in_use), FALSE)`
	columns := m.db.QualifiedColumns(data.EthLogTable)
	columns = append(columns, fmt.Sprintf(topicInAccExpr, topic1)) // topic[1] (agent) in active accounts
	columns = append(columns, fmt.Sprintf(topicInAccExpr, topic2)) // topic[2] (client) in active accounts

	query := fmt.Sprintf(
		`SELECT %s
                          FROM eth_logs
                         WHERE job IS NULL
                               AND NOT ignore`,
		strings.Join(columns, ","),
	)

	var args []interface{}

	maxRetries, err := data.GetUint64Setting(m.db, maxRetryKey)
	if err != nil {
		m.logger.Add("setting", maxRetryKey).Error(err.Error())
		m.errors <- err
		return
	}

	if maxRetries != 0 {
		query += " AND failures <= $1"
		args = append(args, maxRetries)
	}

	query = query + " ORDER BY block_number;"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		m.logger.Add("query", query).Error(err.Error())
		m.errors <- ErrSelectLogsEntries
		return
	}

	for rows.Next() {
		var el data.EthLog
		var forAgent, forClient bool
		pointers := append(el.Pointers(), &forAgent, &forClient)
		if err := rows.Scan(pointers...); err != nil {
			m.logger.Error(err.Error())
			m.errors <- ErrScanRows
			return
		}

		eventHash := el.Topics[0]

		var scheduler funcAndType
		found := false

		if forClient || forAgent {
			scheduler, found = ptcSchedulers[eventHash]
		}

		if !found {
			if m.dappRole == data.RoleAgent {
				if isOfferingRelated(&el) {
					scheduler, found =
						agentOfferingSchedulers[eventHash]
				} else if forAgent {
					scheduler, found =
						agentSchedulers[eventHash]
				}
			} else {
				procedure, ok := clientUpdateCurrentSupplySchedulers[eventHash]
				if ok {
					procedure.f(m, &el, procedure.t)
				}

				if isOfferingRelated(&el) {
					scheduler, found =
						offeringSchedulers[eventHash]
				} else if forClient {
					scheduler, found =
						clientSchedulers[eventHash]
				}
			}
		}

		if !found {
			m.logger.Add("event",
				eventHash.Hex()).Debug("scheduler not" +
				" found for event")
			m.ignoreEvent(&el)
			continue
		}

		scheduler.f(m, &el, scheduler.t)
	}

	if err := rows.Err(); err != nil {
		m.logger.Error(err.Error())
		m.errors <- ErrFetchLogsFromDB
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
		(*Monitor).scheduleAgentChannelCreated,
		data.JobAgentAfterChannelCreate,
	},
	common.HexToHash(eth.EthDigestChannelToppedUp): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobAgentAfterChannelTopUp,
	},
	common.HexToHash(eth.EthChannelCloseRequested): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobAgentAfterUncooperativeCloseRequest,
	},
	common.HexToHash(eth.EthCooperativeChannelClose): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobAgentAfterCooperativeClose,
	},
	common.HexToHash(eth.EthUncooperativeChannelClose): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobAgentAfterUncooperativeClose,
	},
}

var clientSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthDigestChannelCreated): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobClientAfterChannelCreate,
	},
	common.HexToHash(eth.EthDigestChannelToppedUp): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobClientAfterChannelTopUp,
	},
	common.HexToHash(eth.EthChannelCloseRequested): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobClientAfterUncooperativeCloseRequest,
	},
	common.HexToHash(eth.EthCooperativeChannelClose): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobClientAfterCooperativeClose,
	},
	common.HexToHash(eth.EthUncooperativeChannelClose): {
		(*Monitor).scheduleAgentClientChannel,
		data.JobClientAfterUncooperativeClose,
	},
}

var ptcSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthTokenApproval): {
		(*Monitor).scheduleTokenApprove,
		data.JobPreAccountAddBalance,
	},
	common.HexToHash(eth.EthTokenTransfer): {
		(*Monitor).scheduleTokenTransfer,
		"", // determines a job type inside the scheduleTokenTransfer function
	},
}

var agentOfferingSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthOfferingCreated): {
		(*Monitor).scheduleAgentOfferingCreated,
		data.JobAgentAfterOfferingMsgBCPublish,
	},
	common.HexToHash(eth.EthOfferingPoppedUp): {
		(*Monitor).scheduleAgentOfferingCreated,
		data.JobAgentAfterOfferingPopUp,
	},
}

var offeringSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthOfferingCreated): {
		(*Monitor).scheduleClientOfferingCreated,
		data.JobClientAfterOfferingMsgBCPublish,
	},
	common.HexToHash(eth.EthOfferingPoppedUp): {
		(*Monitor).scheduleClientOfferingPopUp,
		data.JobClientAfterOfferingPopUp,
	},
	/* // FIXME: uncomment if monitor should actually delete the offering
	common.HexToHash(eth.EthCOfferingDeleted): {
		(*Monitor).scheduleClient_OfferingDeleted,
		"",
	},
	*/
}

var clientUpdateCurrentSupplySchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthDigestChannelCreated): {
		(*Monitor).scheduleUpdateCurrentSupply,
		data.JobDecrementCurrentSupply,
	},
	common.HexToHash(eth.EthCooperativeChannelClose): {
		(*Monitor).scheduleUpdateCurrentSupply,
		data.JobIncrementCurrentSupply,
	},
	common.HexToHash(eth.EthUncooperativeChannelClose): {
		(*Monitor).scheduleUpdateCurrentSupply,
		data.JobIncrementCurrentSupply,
	},
}

func (m *Monitor) blockNumber(bs []byte, event string) (uint32, error) {
	arg, err := m.pscABI.Events[event].Inputs.NonIndexed().UnpackValues(bs)
	if err != nil {
		m.logger.Add("event",
			m.pscABI.Events[event].Name).Error(err.Error())
		return 0, ErrUnpack
	}

	if len(arg) != m.pscABI.Events[event].Inputs.LengthNonIndexed() {
		m.logger.Add("event",
			m.pscABI.Events[event].Name).Error(
			ErrNumberOfEventArgs.Error())
		return 0, ErrNumberOfEventArgs
	}

	var blockNumber uint32
	var ok bool

	if blockNumber, ok = arg[0].(uint32); !ok {
		m.logger.Error(ErrBlockArgumentType.Error())
		return 0, ErrBlockArgumentType
	}

	return blockNumber, nil
}

// getOpenBlockNumber extracts the Open_block_number field of a given
// channel-related EthLog. Returns false in case it failed, i.e.
// the event has no such field.
func (m *Monitor) getOpenBlockNumber(el *data.EthLog) (uint32, bool, error) {
	bs, err := data.ToBytes(el.Data)
	if err != nil {
		m.logger.Error(err.Error())
		return 0, false, err
	}

	switch el.Topics[0] {
	case common.HexToHash(eth.EthDigestChannelToppedUp):
		blockNumber, err := m.blockNumber(bs, logChannelToppedUp)
		if err != nil {
			return 0, false, err
		}

		return blockNumber, true, nil
	case common.HexToHash(eth.EthChannelCloseRequested):
		blockNumber, err := m.blockNumber(bs, logChannelCloseRequested)
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	case common.HexToHash(eth.EthCooperativeChannelClose):
		blockNumber, err := m.blockNumber(bs,
			logCooperativeChannelClose)
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	case common.HexToHash(eth.EthUncooperativeChannelClose):
		blockNumber, err := m.blockNumber(bs,
			logUnCooperativeChannelClose)
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	case common.HexToHash(eth.EthDigestChannelCreated):
		return 0, false, nil
	}

	return 0, false, ErrUnsupportedTopic
}

func (m *Monitor) findChannelID(el *data.EthLog) string {
	agentAddress := common.BytesToAddress(el.Topics[topic1].Bytes())
	clientAddress := common.BytesToAddress(el.Topics[topic2].Bytes())
	offeringHash := el.Topics[topic3]

	openBlockNumber, hasOpenBlockNumber, err := m.getOpenBlockNumber(el)
	if err != nil {
		return ""
	}

	m.logger.Add("blockNumber", openBlockNumber, "hasBlockNumber",
		hasOpenBlockNumber).Debug("find block number")

	var query string
	args := []interface{}{
		data.FromBytes(offeringHash.Bytes()),
		data.HexFromBytes(agentAddress.Bytes()),
		data.HexFromBytes(clientAddress.Bytes()),
	}
	var row *sql.Row
	if hasOpenBlockNumber {
		query = `SELECT c.id
                           FROM channels AS c, offerings AS o
                          WHERE c.offering = o.id
                                AND o.hash = $1
                                AND c.agent = $2
                                AND c.client = $3
                                AND c.block = $4`
		args = append(args, openBlockNumber)
		row = m.db.QueryRow(query, args...)
	} else {
		query = `SELECT c.id
                           FROM channels AS c, eth_txs AS et
                          WHERE c.id = et.related_id
                                AND et.hash = $1`
		row = m.db.QueryRow(query, el.TxHash)
	}
	var id string
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return ""
		}
		m.logger.Error(err.Error())
		return ""
	}

	return id
}

func (m *Monitor) scheduleUpdateCurrentSupply(el *data.EthLog, jobType string) {
	// HACK: update supply do not mark failures,
	// so we have naive expectation that if eth log has failures,
	// then supply update for it has already been scheduled.
	if el.Failures != 0 {
		return
	}

	var relID string

	// First try to take related id from offering in db.
	topicOffering := el.Topics[3]
	offeringHash := data.FromBytes(topicOffering.Bytes())

	var offering data.Offering
	err := m.db.FindOneTo(&offering, "hash", offeringHash)
	if err == sql.ErrNoRows {
		// Try to take related id from after offering publish job.
		relID = m.offeringPublishJobID(topicOffering)
	} else {
		relID = offering.ID
	}

	if relID == "" {
		m.logger.Add("topicHash", topicOffering.Hex()).Info("not found" +
			" offering by hash to schedule current supply" +
			" update, skipping")
		return
	}

	j := &data.Job{
		Type:        jobType,
		RelatedID:   relID,
		RelatedType: data.JobOffering,
		Data:        []byte("{}"),
		CreatedBy:   data.JobBCMonitor,
	}

	if err := m.queue.Add(nil, j); err != nil {
		m.logger.Error(err.Error())
	}
}

func (m *Monitor) offeringPublishJobID(hash common.Hash) string {
	hashHex := fmt.Sprintf("0x%064v", hash.Hex())
	job := &data.Job{}
	err := m.db.SelectOneTo(job,
		`INNER JOIN eth_logs
		    ON jobs.id=eth_logs.job
		 WHERE eth_logs.topics->>2 = $1
		    AND jobs.type in ($2, $3)`,
		hashHex,
		data.JobAgentAfterOfferingMsgBCPublish,
		data.JobClientAfterOfferingMsgBCPublish,
	)
	if err != nil && err != sql.ErrNoRows {
		m.logger.Error(err.Error())
	}
	return job.ID
}

func (m *Monitor) scheduleTokenApprove(el *data.EthLog, jobType string) {
	addr := common.BytesToAddress(el.Topics[topic1].Bytes())
	addrHash := data.HexFromBytes(addr.Bytes())
	acc := &data.Account{}
	if err := m.db.FindOneTo(acc, "eth_addr", addrHash); err != nil {
		if err == sql.ErrNoRows {
			m.logger.Add("eth_addr",
				addr.String()).Debug("account for" +
				" address not found")
			m.ignoreEvent(el)
			return
		}
		m.logger.Add("eth_addr", addr.String()).Error(err.Error())
		m.ignoreEvent(el)
		return
	}
	j := &data.Job{
		Type:        jobType,
		RelatedID:   acc.ID,
		RelatedType: data.JobAccount,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleTokenTransfer(el *data.EthLog, jobType string) {
	addr1 := common.BytesToAddress(el.Topics[topic1].Bytes())
	addr1Hash := data.HexFromBytes(addr1.Bytes())
	addr2 := common.BytesToAddress(el.Topics[topic2].Bytes())
	addr2Hash := data.HexFromBytes(addr2.Bytes())

	if addr1 == m.pscAddr {
		jobType = data.JobAfterAccountReturnBalance
	} else {
		jobType = data.JobAfterAccountAddBalance
	}

	acc := &data.Account{}
	if err := m.db.SelectOneTo(acc, "where eth_addr=$1 or eth_addr=$2",
		addr1Hash, addr2Hash); err != nil {
		if err == sql.ErrNoRows {
			m.logger.Add("eth_addr1",
				addr1.String(), "eth_addr2",
				addr2.String()).Debug("account for" +
				" address not found")
			m.ignoreEvent(el)
			return
		}
		m.logger.Add("eth_addr1", addr1.String(), "eth_addr2",
			addr2.String()).Info(err.Error())
		m.ignoreEvent(el)
		return
	}
	j := &data.Job{
		Type:        jobType,
		RelatedID:   acc.ID,
		RelatedType: data.JobAccount,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleAgentOfferingCreated(el *data.EthLog,
	jobType string) {
	hashB64 := data.FromBytes(el.Topics[2].Bytes())
	query := `SELECT id
	                  FROM offerings
	                 WHERE hash = $1`

	row := m.db.QueryRow(query, hashB64)
	var id string
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			m.logger.Add("hash",
				el.Topics[2].String()).Debug("offering not" +
				" found with hash")
			m.ignoreEvent(el)
			return
		}
		m.logger.Add("hash", el.Topics[2].String()).Error(err.Error())
		m.ignoreEvent(el)
		return
	}
	j := &data.Job{
		Type:        jobType,
		RelatedID:   id,
		RelatedType: data.JobOffering,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleAgentChannelCreated(el *data.EthLog,
	jobType string) {
	j := &data.Job{
		Type:        jobType,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleAgentClientChannel(el *data.EthLog, jobType string) {
	cid := m.findChannelID(el)
	if cid == "" {
		m.logger.Add("offering",
			el.Topics[topic3].String()).Warn("channel for" +
			" offering does not exist")
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
	query := `SELECT COUNT(*)
                    FROM eth_logs
                   WHERE topics->>0 = $1
                         AND topics->>2 = $2`
	row := m.db.QueryRow(query, "0x"+eth.EthOfferingDeleted,
		offeringHash.Hex())

	var count int
	if err := row.Scan(&count); err != nil {
		m.logger.Add("query", query, "topic",
			"0x"+eth.EthOfferingDeleted).Error(err.Error())
	}

	return count > 0
}

func (m *Monitor) scheduleClientOfferingCreated(el *data.EthLog,
	jobType string) {
	offeringHash := el.Topics[topic2]
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

func (m *Monitor) scheduleClientOfferingPopUp(el *data.EthLog,
	jobType string) {
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
	if j.Data == nil {
		j.Data = []byte("{}")
	}
	err := m.queue.Add(nil, j)
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
		m.logger.Add("id", el.ID, "failures",
			el.Failures).Error(err.Error())
	}
}

func (m *Monitor) updateEventJobID(el *data.EthLog, jobID string) {
	el.JobID = &jobID
	if err := m.db.UpdateColumns(el, "job"); err != nil {
		m.logger.Add("id", el.ID, "job",
			el.JobID).Error(err.Error())
	}
}

func (m *Monitor) ignoreEvent(el *data.EthLog) {
	el.Ignore = true
	if err := m.db.UpdateColumns(el, "ignore"); err != nil {
		m.logger.Add("id", el.ID, "ignore",
			el.Ignore).Error(err.Error())
	}
}
