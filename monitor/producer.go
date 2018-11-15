package monitor

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/util"
)

const (
	logChannelToppedUp           = "LogChannelToppedUp"
	logChannelCloseRequested     = "LogChannelCloseRequested"
	logCooperativeChannelClose   = "LogCooperativeChannelClose"
	logUnCooperativeChannelClose = "LogUnCooperativeChannelClose"
)

// JobsProducers used to bind methods as jobs builder for specific event.
type JobsProducers map[common.Hash]func(*data.JobEthLog) ([]data.Job, error)

func (m *Monitor) agentJobsProducers() JobsProducers {
	return JobsProducers{
		eth.ServiceChannelCreated:            m.agentOnChannelCreated,
		eth.ServiceChannelToppedUp:           m.agentOnChannelToppedUp,
		eth.ServiceChannelCloseRequested:     m.agentOnChannelCloseRequested,
		eth.ServiceOfferingCreated:           m.agentOnOfferingCreated,
		eth.ServiceOfferingDeleted:           m.agentOnOfferingDeleted,
		eth.ServiceOfferingPopedUp:           m.agentOnOfferingPopedUp,
		eth.ServiceCooperativeChannelClose:   m.agentOnCooperativeChannelClose,
		eth.ServiceUnCooperativeChannelClose: m.agentOnUnCooperativeChannelClose,
		eth.TokenApproval:                    m.onTokenApprove,
		eth.TokenTransfer:                    m.onTokenTransfer,
	}
}

func (m *Monitor) clientJobsProducers() JobsProducers {
	return JobsProducers{
		eth.ServiceChannelCreated:            m.clientOnChannelCreated,
		eth.ServiceChannelToppedUp:           m.clientOnChannelToppedUp,
		eth.ServiceChannelCloseRequested:     m.clientOnChannelCloseRequested,
		eth.ServiceOfferingCreated:           m.clientOnOfferingCreated,
		eth.ServiceOfferingDeleted:           m.clientOnOfferingDeleted,
		eth.ServiceOfferingPopedUp:           m.clientOnOfferingPopedUp,
		eth.ServiceCooperativeChannelClose:   m.clientOnCooperativeChannelClose,
		eth.ServiceUnCooperativeChannelClose: m.clientOnUnCooperativeChannelClose,
		eth.TokenApproval:                    m.onTokenApprove,
		eth.TokenTransfer:                    m.onTokenTransfer,
	}
}

func (m *Monitor) agentOnChannelCreated(l *data.JobEthLog) ([]data.Job, error) {
	offering := l.Topics[3]
	oid := m.findOfferingID(offering)
	if oid == "" {
		return nil, nil
	}
	return m.produceCommon(l, util.NewUUID(), data.JobChannel,
		data.JobAgentAfterChannelCreate)
}

func (m *Monitor) agentOnChannelToppedUp(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterChannelTopUp)
}

func (m *Monitor) agentOnChannelCloseRequested(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterUncooperativeCloseRequest)
}

func (m *Monitor) agentOnCooperativeChannelClose(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterCooperativeClose)
}

func (m *Monitor) agentOnUnCooperativeChannelClose(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterUncooperativeClose)
}

func (m *Monitor) agentOnOfferingCreated(l *data.JobEthLog) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingMsgBCPublish)
}

func (m *Monitor) agentOnOfferingPopedUp(l *data.JobEthLog) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingPopUp)
}

func (m *Monitor) agentOnOfferingDeleted(l *data.JobEthLog) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingDelete)
}

func (m *Monitor) clientOnChannelCreated(l *data.JobEthLog) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterChannelCreate)
	if err != nil {
		return nil, err
	}

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobDecrementCurrentSupply)

	if err != nil {
		return nil, err
	}

	return append(jobs, updateJobs...), nil
}

func (m *Monitor) clientOnChannelToppedUp(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobClientAfterChannelTopUp)
}

func (m *Monitor) clientOnChannelCloseRequested(l *data.JobEthLog) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobClientAfterUncooperativeCloseRequest)
}

func (m *Monitor) clientOnOfferingCreated(l *data.JobEthLog) ([]data.Job, error) {
	return m.produceCommon(l, util.NewUUID(), data.JobOffering,
		data.JobClientAfterOfferingMsgBCPublish)
}

func (m *Monitor) clientOnOfferingDeleted(l *data.JobEthLog) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobClientAfterOfferingDelete)
}

func (m *Monitor) clientOnOfferingPopedUp(l *data.JobEthLog) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobClientAfterOfferingPopUp)
}

func (m *Monitor) clientOnCooperativeChannelClose(l *data.JobEthLog) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterCooperativeClose)
	if err != nil {
		return nil, err
	}

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobIncrementCurrentSupply)

	if err != nil {
		return nil, err
	}

	return append(jobs, updateJobs...), nil
}

func (m *Monitor) clientOnUnCooperativeChannelClose(l *data.JobEthLog) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterUncooperativeClose)
	if err != nil {
		return nil, err
	}

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobIncrementCurrentSupply)

	if err != nil {
		return nil, err
	}

	return append(jobs, updateJobs...), nil
}

func (m *Monitor) onExistingChannelEvent(l *data.JobEthLog, jtype string) ([]data.Job, error) {
	cid := m.findChannelID(l)
	if cid == "" {
		m.logger.Add("ethereumLog", *l).Warn("channel not found")
		return nil, nil
	}
	return m.produceCommon(l, cid, data.JobChannel, jtype)
}

func (m *Monitor) onOfferingRelatedEvent(l *data.JobEthLog, jtype string) ([]data.Job, error) {
	oid := m.findOfferingID(l.Topics[2])
	if oid == "" {
		return nil, nil
	}
	return m.produceCommon(l, oid, data.JobOffering, jtype)
}

func (m *Monitor) onTokenApprove(l *data.JobEthLog) ([]data.Job, error) {
	addr := common.BytesToAddress(l.Topics[1].Bytes())
	addrHash := data.HexFromBytes(addr.Bytes())
	acc := &data.Account{}
	if err := m.db.FindOneTo(acc, "eth_addr", addrHash); err != nil {
		logger := m.logger.Add("eth_addr", addr.String())
		if err == sql.ErrNoRows {
			logger.Debug("account for address not found")
			return nil, nil
		}
		logger.Error(err.Error())
		return nil, nil
	}
	return m.produceCommon(l, acc.ID, data.JobAccount, data.JobPreAccountAddBalance)
}

func (m *Monitor) onTokenTransfer(l *data.JobEthLog) ([]data.Job, error) {
	addr1 := common.BytesToAddress(l.Topics[1].Bytes())
	addr1Hash := data.HexFromBytes(addr1.Bytes())
	addr2 := common.BytesToAddress(l.Topics[2].Bytes())
	addr2Hash := data.HexFromBytes(addr2.Bytes())

	var jtype string
	if addr1 == m.pscAddr {
		jtype = data.JobAfterAccountReturnBalance
	} else {
		jtype = data.JobAfterAccountAddBalance
	}

	acc := &data.Account{}
	if err := m.db.SelectOneTo(acc, "WHERE eth_addr=$1 OR eth_addr=$2",
		addr1Hash, addr2Hash); err != nil {
		logger := m.logger.Add("eth_addr1", addr1.String(), "eth_addr2",
			addr2.String())
		if err == sql.ErrNoRows {
			logger.Debug("account for address not found")
			return nil, nil
		}
		logger.Error(err.Error())
		return nil, nil
	}

	return m.produceCommon(l, acc.ID, data.JobAccount, jtype)
}

func (m *Monitor) findChannelID(el *data.JobEthLog) string {
	agentAddress := common.BytesToAddress(el.Topics[1].Bytes())
	clientAddress := common.BytesToAddress(el.Topics[2].Bytes())
	offeringHash := el.Topics[3]

	openBlockNumber, hasOpenBlockNumber, err := m.getOpenBlockNumber(el)
	if err != nil {
		return ""
	}

	m.logger.Add("blockNumber", openBlockNumber, "hasBlockNumber",
		hasOpenBlockNumber).Debug("find block number")

	var query string
	args := []interface{}{
		data.HexFromBytes(offeringHash.Bytes()),
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

func (m *Monitor) findOfferingID(topic common.Hash) string {
	hashHex := data.HexFromBytes(topic.Bytes())
	offering := &data.Offering{}
	err := m.db.FindOneTo(offering, "hash", hashHex)
	if err != nil {
		l := m.logger.Add("hash", topic.String())
		if err == sql.ErrNoRows {
			l.Debug("offering not found with hash")
			return ""
		}
		l.Error(err.Error())
	}
	return offering.ID
}

func (m *Monitor) updateCurrentSupplyJobs(l *data.JobEthLog, jtype string) ([]data.Job, error) {
	offering := l.Topics[3]
	oid := m.findOfferingID(offering)
	if oid == "" {
		oid = m.offeringCreateJobRelatedID(offering)
	}

	if oid == "" {
		return nil, nil
	}

	return m.produceCommon(l, oid, data.JobOffering, jtype)
}

func (m *Monitor) produceCommon(l *data.JobEthLog, rid, rtype, jtype string) ([]data.Job, error) {
	jdata, err := json.Marshal(&data.JobData{EthLog: l})
	if err != nil {
		return nil, err
	}
	return []data.Job{
		{
			Data:        jdata,
			RelatedID:   rid,
			RelatedType: rtype,
			Type:        jtype,
		},
	}, nil
}

// getOpenBlockNumber extracts the Open_block_number field of a given
// channel-related EthLog. Returns false in case it failed, i.e.
// the event has no such field.
func (m *Monitor) getOpenBlockNumber(
	el *data.JobEthLog) (block uint32, ok bool, err error) {
	var event string
	switch el.Topics[0] {
	case eth.ServiceChannelToppedUp:
		event = logChannelToppedUp
	case eth.ServiceChannelCloseRequested:
		event = logChannelCloseRequested
	case eth.ServiceCooperativeChannelClose:
		event = logCooperativeChannelClose
	case eth.ServiceUnCooperativeChannelClose:
		event = logUnCooperativeChannelClose
	default:
		return
	}

	if event != "" {
		block, err = m.blockNumber(el.Data,
			logChannelToppedUp)
		ok = true
	}
	return
}

func (m *Monitor) offeringCreateJobRelatedID(hash common.Hash) string {
	hashHex := fmt.Sprintf("0x%064v", hash.Hex())
	job := &data.Job{}
	err := m.db.SelectOneTo(job,
		`WHERE data->'ethereum_logs'->'topics'->>2 = $1
		    AND jobs.type in ($2, $3)`,
		hashHex,
		data.JobClientAfterOfferingMsgBCPublish,
		data.JobClientAfterOfferingPopUp,
	)
	if err != nil && err != sql.ErrNoRows {
		m.logger.Error(err.Error())
	}
	return job.RelatedID
}

func (m *Monitor) blockNumber(bs []byte, event string) (uint32, error) {
	logger := m.logger.Add("method", "blockNumber")

	arg, err := m.pscABI.Events[event].Inputs.NonIndexed().UnpackValues(bs)
	if err != nil {
		logger.Add("event",
			m.pscABI.Events[event].Name).Error(err.Error())
		return 0, ErrFailedToUnpack
	}

	if len(arg) != m.pscABI.Events[event].Inputs.LengthNonIndexed() {
		return 0, ErrWrongNumberOfEventArgs
	}

	var blockNumber uint32
	var ok bool

	if blockNumber, ok = arg[0].(uint32); !ok {
		logger.Error(ErrWrongBlockArgumentType.Error())
		return 0, ErrWrongBlockArgumentType
	}

	return blockNumber, nil
}
