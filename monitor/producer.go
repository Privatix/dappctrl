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

// JobsProducers used to bind methods as jobs builder for specific event.
// First argument is log to proccess.
// Second argument is jobs produced but not yet created in db. It is used to
// produce jobs that might have relateds in current produced set.
type JobsProducers map[common.Hash]func(*data.JobEthLog, []data.Job) ([]data.Job, error)

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

func (m *Monitor) agentOnChannelCreated(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	offering := l.Topics[3]
	oid := m.findOfferingID(offering)
	if oid == "" {
		return nil, nil
	}
	return m.produceCommon(l, util.NewUUID(), data.JobChannel,
		data.JobAgentAfterChannelCreate)
}

func (m *Monitor) agentOnChannelToppedUp(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterChannelTopUp)
}

func (m *Monitor) agentOnChannelCloseRequested(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterUncooperativeCloseRequest)
}

func (m *Monitor) agentOnCooperativeChannelClose(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterCooperativeClose)
}

func (m *Monitor) agentOnUnCooperativeChannelClose(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobAgentAfterUncooperativeClose)
}

func (m *Monitor) agentOnOfferingCreated(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingMsgBCPublish)
}

func (m *Monitor) agentOnOfferingPopedUp(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingPopUp)
}

func (m *Monitor) agentOnOfferingDeleted(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobAgentAfterOfferingDelete)
}

func (m *Monitor) clientOnChannelCreated(l *data.JobEthLog, producingJobs []data.Job) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterChannelCreate)
	if err != nil {
		return nil, err
	}

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobDecrementCurrentSupply, producingJobs)

	if err != nil {
		return nil, err
	}

	return append(jobs, updateJobs...), nil
}

func (m *Monitor) clientOnChannelToppedUp(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobClientAfterChannelTopUp)
}

func (m *Monitor) clientOnChannelCloseRequested(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onExistingChannelEvent(l, data.JobClientAfterUncooperativeCloseRequest)
}

func (m *Monitor) clientOnOfferingCreated(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.produceCommon(l, util.NewUUID(), data.JobOffering,
		data.JobClientAfterOfferingMsgBCPublish)
}

func (m *Monitor) clientOnOfferingDeleted(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	return m.onOfferingRelatedEvent(l, data.JobClientAfterOfferingDelete)
}

func (m *Monitor) clientOnOfferingPopedUp(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
	oid := m.findOfferingID(l.Topics[2])
	if oid == "" {
		oid = util.NewUUID()
	}
	return m.produceCommon(l, oid, data.JobOffering, data.JobClientAfterOfferingPopUp)
}

func (m *Monitor) clientOnCooperativeChannelClose(l *data.JobEthLog, producingJobs []data.Job) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterCooperativeClose)
	if err != nil {
		return nil, err
	}

	rateJob, err := m.rateJob(l, data.ClosingCoop)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, rateJob)

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobIncrementCurrentSupply, producingJobs)
	if err != nil {
		return nil, err
	}

	if len(updateJobs) > 0 {
		created, err := m.incrementSupplyAlreadyCreated(l)
		if err != nil {
			return nil, err
		}
		if created {
			return jobs, nil
		}
	}

	return append(jobs, updateJobs...), nil
}

func (m *Monitor) clientOnUnCooperativeChannelClose(l *data.JobEthLog, producingJobs []data.Job) ([]data.Job, error) {
	jobs, err := m.onExistingChannelEvent(l, data.JobClientAfterUncooperativeClose)
	if err != nil {
		return nil, err
	}

	rateJob, err := m.rateJob(l, data.ClosingCoop)
	if err != nil {
		return nil, err
	}
	jobs = append(jobs, rateJob)

	updateJobs, err := m.updateCurrentSupplyJobs(l, data.JobIncrementCurrentSupply, producingJobs)
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

func (m *Monitor) onTokenApprove(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
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

func (m *Monitor) onTokenTransfer(l *data.JobEthLog, _ []data.Job) ([]data.Job, error) {
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
		hasOpenBlockNumber).Debug("find block number in Ethereum log")

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

func (m *Monitor) updateCurrentSupplyJobs(
	l *data.JobEthLog, jtype string, producingJobs []data.Job) ([]data.Job, error) {
	offering := l.Topics[3]
	oid := m.findOfferingID(offering)
	if oid == "" {
		oid = m.offeringCreateJobRelatedID(offering)
		if oid == "" {
			oid = m.findByHashIn(offering, producingJobs)
		}
		if oid == "" {
			return nil, nil
		}
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

func (m *Monitor) rateJob(l *data.JobEthLog, closingType string) (data.Job, error) {
	// Find cooperative closing channel block number.
	inputs := m.pscABI.Events[logCooperativeChannelClose].Inputs
	if closingType == data.ClosingUncoop {
		inputs = m.pscABI.Events[logUnCooperativeChannelClose].Inputs
	}
	args, err := inputs.UnpackValues(l.Data)
	if err != nil {
		m.logger.Error(err.Error())
		return data.Job{}, ErrInternal
	}
	jdata, err := json.Marshal(&data.JobRecordClosingData{
		Rec: &data.Closing{
			ID:      util.NewUUID(),
			Type:    closingType,
			Agent:   data.HexFromBytes(common.BytesToAddress(l.Topics[1].Bytes()).Bytes()),
			Client:  data.HexFromBytes(common.BytesToAddress(l.Topics[2].Bytes()).Bytes()),
			Balance: args[1].(uint64),
			Block:   args[0].(uint32),
		},
		UpdateRatings: m.updateRating(),
	})
	if err != nil {
		m.logger.Error(err.Error())
		return data.Job{}, ErrInternal
	}
	return data.Job{
		Data:        jdata,
		RelatedID:   util.NewUUID(),
		RelatedType: data.JobChannel,
		Type:        data.JobClientRecordClosing,
	}, nil
}

// updateRating called on channel closing event.
// Returns true if monitor received rateAfter closings after last ratings calculation.
func (m *Monitor) updateRating() bool {
	m.closingsCount++
	if m.closingsCount >= m.rateAfter {
		m.closingsCount = 0
		return true
	}
	return false
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
	hashHex := hashToHex(hash)
	job := &data.Job{}
	err := m.db.SelectOneTo(job,
		`WHERE data->'ethereumLog'->'topics'->>2 = $1
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

func (m *Monitor) findByHashIn(hash common.Hash, jobs []data.Job) string {
	for _, job := range jobs {
		if job.Type == data.JobClientAfterOfferingMsgBCPublish ||
			job.Type == data.JobClientAfterOfferingPopUp {
			jdata := &data.JobData{EthLog: &data.JobEthLog{}}

			err := json.Unmarshal(job.Data, jdata)
			if err != nil {
				m.logger.Error("failed to unmarshal ethLog while searching" +
					"for client after publish job")
			}

			if err == nil {

				topics := jdata.EthLog.Topics
				if len(topics) >= 3 && topics[2] == hash {
					return job.RelatedID
				}
			}
		}
	}
	return ""
}

func (m *Monitor) incrementSupplyAlreadyCreated(l *data.JobEthLog) (bool, error) {
	if l.Topics[0] != eth.ServiceCooperativeChannelClose && len(l.Topics) != 4 {
		return false, ErrWrongNumberOfEventArgs
	}
	// Find cooperative closing channel block number.
	inputs := m.pscABI.Events[logUnCooperativeChannelClose].Inputs
	args, err := inputs.UnpackValues(l.Data)
	if err != nil {
		m.logger.Error(err.Error())
		return false, ErrInternal
	}
	closingChannelblock := args[0].(uint32)

	event := hashToHex(eth.ServiceUnCooperativeChannelClose)
	agent := hashToHex(l.Topics[1])
	client := hashToHex(l.Topics[2])
	offering := hashToHex(l.Topics[3])
	jobsStructs, err := m.db.SelectAllFrom(data.JobTable,
		`WHERE data->'ethereumLog'->'topics'->>0 = $1
			AND data->'ethereumLog'->'topics'->>1 = $2
			AND data->'ethereumLog'->'topics'->>2 = $3
			AND data->'ethereumLog'->'topics'->>3 = $4
			AND type = $5`,
		event, agent, client, offering, data.JobIncrementCurrentSupply,
	)
	if err != nil {
		m.logger.Error(err.Error())
		return false, ErrInternal
	}
	for _, item := range jobsStructs {
		job := item.(*data.Job)
		jdata := &data.JobData{}
		err := json.Unmarshal(job.Data, jdata)
		if err != nil {
			m.logger.Error(err.Error())
			return false, ErrInternal
		}
		inputs := m.pscABI.Events[logCooperativeChannelClose].Inputs
		args, err := inputs.UnpackValues(jdata.EthLog.Data)
		if err != nil {
			m.logger.Error(err.Error())
			return false, ErrInternal
		}
		if args[0].(uint32) == closingChannelblock {
			return true, nil
		}
	}
	return false, nil
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

func hashToHex(hash common.Hash) string {
	ret := hash.Hex()
	if len(ret) < 2 || ret[:2] != "0x" {
		ret = fmt.Sprintf("0x%064v", hash.Hex())
	}
	return ret
}
