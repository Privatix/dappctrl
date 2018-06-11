package monitor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

const maxRetryKey = "eth.event.maxretry"

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
func (m *Monitor) schedule(ctx context.Context, timeout int64,
	errCh chan error) {
	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(timeout)*time.Second)
	defer cancel()

	// TODO: Move this logic into a database view? The query is just supposed to
	// append two boolean columns calculated based on topics: whether the
	// event is for agent, and the same for client.
	//
	// eth_logs.topics is a json array with '0x0..0deadbeef' encoding
	// of addresses, whereas accounts.eth_addr is a base64 encoding of
	// raw bytes of addresses.
	// The encode-decode-substr is there to convert from one to another.
	// coalesce() converts null into false for the case when topics->>n
	// does not exist.
	topicInAccExpr := `COALESCE(TRANSLATE(encode(decode(substr(topics->>%d, 27), 'hex'), 'base64'), '+/', '-_') IN (SELECT eth_addr FROM accounts WHERE in_use), FALSE)`
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
		m.errWrapper(ctx, err)
	}

	if maxRetries != 0 {
		query += " AND failures <= $1"
		args = append(args, maxRetries)
	}

	query = query + " ORDER BY block_number;"

	rows, err := m.db.Query(query, args...)
	if err != nil {
		m.errWrapper(ctx,
			fmt.Errorf("failed to select log entries: %v", err))
		return
	}

	for rows.Next() {
		var el data.EthLog
		var forAgent, forClient bool
		pointers := append(el.Pointers(), &forAgent, &forClient)
		if err := rows.Scan(pointers...); err != nil {
			m.errWrapper(ctx,
				fmt.Errorf("failed to scan the selected log"+
					" entries: %v", err))
			return
		}

		eventHash := el.Topics[0]

		var scheduler funcAndType
		found := false
		switch {
		case forAgent:
			scheduler, found = agentSchedulers[eventHash]
			if !found {
				scheduler, found = accountSchedulers[eventHash]
			}
			// TODO: uncomment when client jobs will be implemented
			/*
				case forClient:
					scheduler, found = clientSchedulers[eventHash]
					if !found {
						scheduler, found = accountSchedulers[eventHash]
					}
				case isOfferingRelated(&el):
					scheduler, found = offeringSchedulers[eventHash]
			*/
		}

		if !found {
			m.logger.Debug("scheduler not found for event %s",
				eventHash.Hex())
			m.ignoreEvent(&el)
			continue
		}

		scheduler.f(m, &el, scheduler.t)
	}

	if err := rows.Err(); err != nil {
		m.errWrapper(ctx,
			fmt.Errorf("failed to fetch the next selected log"+
				" entry: %v", err))
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

var accountSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthTokenApproval): {
		(*Monitor).scheduleTokenApprove,
		data.JobPreAccountAddBalance,
	},
	common.HexToHash(eth.EthTokenTransfer): {
		(*Monitor).scheduleTokenTransfer,
		data.JobAfterAccountAddBalance,
	},
	// TODO: return balance schedulers
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
	common.HexToHash(eth.EthOfferingCreated): {
		(*Monitor).scheduleAgentOfferingCreated,
		data.JobAgentAfterOfferingMsgBCPublish,
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

var offeringSchedulers = map[common.Hash]funcAndType{
	common.HexToHash(eth.EthOfferingCreated): {
		(*Monitor).scheduleClientOfferingCreated,
		data.JobClientAfterOfferingMsgBCPublish,
	},
	common.HexToHash(eth.EthOfferingPoppedUp): {
		(*Monitor).scheduleClientOfferingCreated,
		data.JobClientAfterOfferingMsgBCPublish,
	},
	/* // FIXME: uncomment if monitor should actually delete the offering
	common.HexToHash(eth.EthCOfferingDeleted): {
		(*Monitor).scheduleClient_OfferingDeleted,
		"",
	},
	*/
}

func (m *Monitor) blockNumber(bs []byte, event string) (uint32, error) {
	arg, err := m.pscABI.Events[event].
		Inputs.NonIndexed().UnpackValues(bs)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack arguments: %v", err)
	}

	if len(arg) != m.pscABI.Events[event].
		Inputs.LengthNonIndexed() {
		return 0, fmt.Errorf("wrong number of event arguments")
	}

	var blockNumber uint32
	var ok bool

	if blockNumber, ok = arg[0].(uint32); !ok {
		return 0, fmt.Errorf("wrong type argument of OpenBlockNumber")
	}

	return blockNumber, nil
}

// getOpenBlockNumber extracts the Open_block_number field of a given
// channel-related EthLog. Returns false in case it failed, i.e.
// the event has no such field.
func (m *Monitor) getOpenBlockNumber(el *data.EthLog) (uint32, bool, error) {
	bs, err := data.ToBytes(el.Data)
	if err != nil {
		return 0, false, err
	}

	switch el.Topics[0] {
	case common.HexToHash(eth.EthDigestChannelToppedUp):
		blockNumber, err := m.blockNumber(bs,
			"LogChannelToppedUp")
		if err != nil {
			return 0, false, err
		}

		return blockNumber, true, nil
	case common.HexToHash(eth.EthChannelCloseRequested):
		blockNumber, err := m.blockNumber(bs,
			"LogChannelCloseRequested")
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	case common.HexToHash(eth.EthCooperativeChannelClose):
		blockNumber, err := m.blockNumber(bs,
			"LogCooperativeChannelClose")
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	case common.HexToHash(eth.EthUncooperativeChannelClose):
		blockNumber, err := m.blockNumber(bs,
			"LogUnCooperativeChannelClose")
		if err != nil {
			return 0, false, err
		}
		return blockNumber, true, nil
	}

	return 0, false, fmt.Errorf("unsupported topic")
}

func (m *Monitor) findChannelID(el *data.EthLog) string {
	agentAddress := common.BytesToAddress(el.Topics[topic1].Bytes())
	clientAddress := common.BytesToAddress(el.Topics[topic2].Bytes())
	offeringHash := el.Topics[topic3]

	openBlockNumber, haveOpenBlockNumber, err := m.getOpenBlockNumber(el)
	if err != nil {
		m.logger.Warn(err.Error())
		return ""
	}

	m.logger.Info("bn = %d, hbn = %t", openBlockNumber,
		haveOpenBlockNumber)

	var query string
	args := []interface{}{
		data.FromBytes(offeringHash.Bytes()),
		data.FromBytes(agentAddress.Bytes()),
		data.FromBytes(clientAddress.Bytes()),
	}
	if haveOpenBlockNumber {
		query = `SELECT c.id
                           FROM channels AS c, offerings AS o
                          WHERE c.offering = o.id
                                AND o.hash = $1
                                AND c.agent = $2
                                AND c.client = $3
                                AND c.block = $4`
		args = append(args, openBlockNumber)
	} else {
		query = `SELECT c.id
                           FROM channels AS c, offerings AS o, eth_txs AS et
                          WHERE c.offering = o.id
                                AND o.hash = $1
                                AND c.agent = $2
                                AND c.client = $3
                                AND et.hash = $4`
		args = append(args, el.TxHash)
	}
	row := m.db.QueryRow(query, args...)

	var id string
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return ""
		}
		m.logger.Error("failed to scan row %s", err)
		return ""
	}

	return id
}

func (m *Monitor) scheduleTokenApprove(el *data.EthLog, jobType string) {
	addr := common.BytesToAddress(el.Topics[topic1].Bytes())
	addrHash := data.FromBytes(addr.Bytes())
	acc := &data.Account{}
	if err := m.db.FindOneTo(acc, "eth_addr", addrHash); err != nil {
		if err == sql.ErrNoRows {
			m.logger.Debug("account not found for addr %s",
				el.Topics[1].Hex())
			m.ignoreEvent(el)
			return
		}
		m.logger.Error("failed to find account: %v", err)
		m.ignoreEvent(el)
		return
	}
	amountBytes, err := data.ToBytes(el.Data)
	if err != nil {
		m.logger.Error("failed to decode eth log data: %v", err)
		m.ignoreEvent(el)
		return
	}
	balanceData := &data.JobBalanceData{
		Amount: uint(big.NewInt(0).SetBytes(amountBytes).Uint64()),
	}
	dataEncoded, err := json.Marshal(balanceData)
	if err != nil {
		m.logger.Error("failed to marshal balance data: %v", err)
		m.ignoreEvent(el)
		return
	}
	j := &data.Job{
		Type:        jobType,
		RelatedID:   acc.ID,
		RelatedType: data.JobAccount,
		Data:        dataEncoded,
	}

	m.scheduleCommon(el, j)
}

func (m *Monitor) scheduleTokenTransfer(el *data.EthLog, jobType string) {
	addr1 := common.BytesToAddress(el.Topics[topic1].Bytes())
	addr1Hash := data.FromBytes(addr1.Bytes())
	addr2 := common.BytesToAddress(el.Topics[topic2].Bytes())
	addr2Hash := data.FromBytes(addr2.Bytes())

	acc := &data.Account{}
	if err := m.db.SelectOneTo(acc, "where eth_addr=$1 or eth_addr=$2", addr1Hash,
		addr2Hash); err != nil {
		if err == sql.ErrNoRows {
			m.logger.Info("account not found for addr %s",
				el.Topics[1].Hex())
			m.ignoreEvent(el)
			return
		}
		m.logger.Info("failed to find account: %v", err)
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
			m.logger.Debug("offering not found with hash %s",
				el.Topics[2].Hex())
			m.ignoreEvent(el)
			return
		}
		m.logger.Error("failed to scan row %s", err)
		m.ignoreEvent(el)
		return
	}
	j := &data.Job{
		Type:        jobType,
		RelatedID:   id,
		RelatedType: data.JobOfferring,
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
		m.logger.Warn("channel for offering %s does not exist",
			el.Topics[topic3].Hex())
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
		m.logger.Error("failed to scan row %s", err)
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
		RelatedType: data.JobOfferring,
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
	if j.Data == nil {
		j.Data = []byte("{}")
	}
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
		m.logger.Error("failed to update failure counter"+
			" of an event: %v", err)
	}
}

func (m *Monitor) updateEventJobID(el *data.EthLog, jobID string) {
	el.JobID = &jobID
	if err := m.db.UpdateColumns(el, "job"); err != nil {
		m.logger.Error("failed to update job_id of an event"+
			" to %s: %v", jobID, err)
	}
}

func (m *Monitor) ignoreEvent(el *data.EthLog) {
	el.Ignore = true
	if err := m.db.UpdateColumns(el, "ignore"); err != nil {
		m.logger.Error("failed to ignore an event: %v", err)
	}
}
