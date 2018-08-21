package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

const (
	channelTerminate   = "terminate"
	channelPause       = "pause"
	channelResume      = "resume"
	clientChannelClose = "close"
)

var (
	channelsGetParams = []queryParam{
		{Name: "id", Field: "id"},
		{Name: "channelStatus", Field: "channel_status"},
		{Name: "serviceStatus", Field: "service_status"},
	}

	clientStatusFilter   = `WHERE id = $1 AND channels.agent NOT IN (SELECT eth_addr FROM accounts)`
	clientChannelsFilter = `WHERE channels.agent NOT IN (SELECT eth_addr FROM accounts)`
)

type chanStatusBlock struct {
	ServiceStatus   string  `json:"serviceStatus"`
	ChannelStatus   string  `json:"channelStatus"`
	LastChanged     *string `json:"lastChanged"`
	MaxInactiveTime uint64  `json:"maxInactiveTime"`
}

type jobBlock struct {
	ID        string `json:"id"`
	Type      string `json:"jobtype"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type usageBlock struct {
	Current  uint64 `json:"current"`
	MaxUsage uint64 `json:"maxUsage"`
	Unit     string `json:"unit"`
	Cost     uint64 `json:"cost"`
}

type usage struct {
	unitType    string
	secUsage    uint64
	unitsUsage  uint64
	costSeconds uint64
	costUnits   uint64
}

type respGetClientChan struct {
	ID       string `json:"id"`
	Agent    string `json:"agent"`
	Client   string `json:"client"`
	Offering string `json:"offering"`
	Deposit  uint64 `json:"deposit"`

	ChStat chanStatusBlock `json:"channelStatus"`
	Job    jobBlock        `json:"job"`
	Usage  usageBlock      `json:"usage"`
}

// handleChannels calls appropriate handler by scanning incoming request.
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(channelsPath, r.URL.Path); id != "" {
		if r.Method == "GET" {
			s.handleGetChannelStatus(w, r, id)
			return
		}
		if r.Method == "PUT" {
			s.handlePutChannelStatus(w, r, id)
			return
		}
	} else {
		if r.Method == "GET" {
			s.handleGetChannels(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleChannels calls appropriate handler by scanning incoming request.
func (s *Server) handleClientChannels(w http.ResponseWriter, r *http.Request) {
	if id := idFromStatusPath(clientChannelsPath, r.URL.Path); id != "" {
		if r.Method == "GET" {
			s.handleGetClientChannelStatus(w, r, id)
			return
		}
		if r.Method == "PUT" {
			s.handlePutClientChannelStatus(w, r, id)
			return
		}
	} else {
		if r.Method == "GET" {
			s.handleGetClientChannels(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGetChannels replies with all channels or a channel by id
// available to the agent.
func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	s.handleGetResources(w, r, &getConf{
		Params: channelsGetParams,

		View:         data.ChannelTable,
		FilteringSQL: `channels.agent IN (SELECT eth_addr FROM accounts)`,
	})
}

func (s *Server) filter(conds []string) (constraints string) {
	if len(conds) > 0 {
		for k, v := range conds {
			conds[k] = "channels." + v
		}

		constraints = " AND " + strings.Join(conds, " AND ")
	}

	return constraints
}

// ethAddrFromHex returns ethereum's address on string format from hex encoded string
// if the address is not valid, it returns an empty string
func ethAddrFromHex(addr string) string {
	ethAddr, err := data.HexToAddress(addr)
	if err != nil {
		ethAddr = common.Address{}
	}
	return ethAddr.String()
}

func formatTimeStr(tm *string) *string {
	if tm != nil && *tm != "" {
		*tm = singleTimeFormatFromStr(*tm)
		return tm
	}
	return new(string)
}

func usageCalc(u *usageBlock, params *usage) {
	if params.unitType == data.UnitScalar {
		u.Cost = params.costUnits
		u.Current = params.unitsUsage
	} else if params.unitType == data.UnitSeconds {
		u.Cost = params.costSeconds
		u.Current = params.secUsage
	}
}

func (s *Server) parseTime(logger log.Logger,
	w http.ResponseWriter, value string) (ti time.Time, err error) {
	ti, err = time.Parse(timeFormat, value)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to parse time: %v", err))
		s.replyUnexpectedErr(logger, w)
	}
	return
}

func (s *Server) getClientChannelsItems(w http.ResponseWriter, query string,
	args []interface{}) (resp []*respGetClientChan, err error) {
	logger := s.logger.Add("method", "getClientChannelsItems",
		"query", query)
	resp = []*respGetClientChan{}

	channels := make(map[string]*respGetClientChan)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to select: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		item := new(respGetClientChan)
		i := new(usage)

		if err = rows.Scan(&item.ID, &item.Agent, &item.Client,
			&item.Offering, &item.Deposit,
			&item.ChStat.ServiceStatus, &item.ChStat.ChannelStatus,
			&item.ChStat.LastChanged, &item.ChStat.MaxInactiveTime,
			&item.Job.ID, &item.Job.Type, &item.Job.Status,
			&item.Job.CreatedAt, &i.secUsage,
			&i.unitsUsage, &item.Usage.MaxUsage, &i.unitType,
			&item.Usage.Unit, &i.costSeconds,
			&i.costUnits); err != nil {
			logger.Warn(fmt.Sprintf("failed to scan rows: %v", err))
			s.replyUnexpectedErr(logger, w)
			return
		}

		if item.Job.Status != data.JobDone {
			continue
		}

		// time formatting
		item.ChStat.LastChanged = formatTimeStr(item.ChStat.LastChanged)
		formatTimeStr(&item.Job.CreatedAt)

		// client ETH address conversion
		item.Client = ethAddrFromHex(item.Client)
		item.Agent = ethAddrFromHex(item.Agent)

		// usage calculation
		usageCalc(&item.Usage, i)

		if old, ok := channels[item.ID]; ok {
			var oldTime time.Time
			var newTime time.Time

			oldTime, err = s.parseTime(logger, w, old.Job.CreatedAt)
			if err != nil {
				return
			}

			newTime, err = s.parseTime(logger, w, item.Job.CreatedAt)
			if err != nil {
				return
			}

			if newTime.After(oldTime) {
				channels[item.ID] = item
			}
			continue
		}

		channels[item.ID] = item
	}
	if err = rows.Err(); err != nil {
		logger.Warn(fmt.Sprintf("failed to rows iteration: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	for _, ch := range channels {
		resp = append(resp, ch)
	}

	return resp, nil
}

// handleGetClientChannels replies with all client channels or a channel by id
// available to the client.
func (s *Server) handleGetClientChannels(w http.ResponseWriter,
	r *http.Request) {
	logger := s.logger.Add("method", "handleGetClientChannels")

	// Result 20 fields: id, agent, client, offering, Deposit, service_status, channel_status,
	// last_changed, max_inactive_time_sec, job_id, job_type, job_status, job_created_at,
	// sec_usage, units_usage, max_usage, unit_type, unit_name, cost_seconds, cost_units
	queryHeader := `
		SELECT channels.id, channels.agent, channels.client, channels.offering,
                       channels.total_deposit AS Deposit, channels.service_status,
                       channels.channel_status, channels.service_changed_time AS last_changed,
                       COALESCE(offer.max_inactive_time_sec, 0),
                       job.id AS job_id, job.type AS job_type,
                       job.status AS job_status, job.created_at AS job_created_at,
                       COALESCE(SUM(ses.seconds_consumed), 0) AS sec_usage,
                       COALESCE(SUM(ses.units_used), 0) AS units_usage,
                       GREATEST(COALESCE(((channels.total_deposit - offer.setup_price) / offer.unit_price), 0), 0) AS max_usage,
                       offer.unit_type, offer.unit_name,
                       COALESCE(offer.setup_price + COALESCE(SUM(ses.seconds_consumed), 0) * offer.unit_price, 0) AS cost_seconds,
                       COALESCE(offer.setup_price + COALESCE(SUM(ses.units_used), 0) * offer.unit_price, 0) AS cost_units
                  FROM channels
                       LEFT JOIN sessions ses
                       ON channels.id = ses.channel

                       LEFT JOIN offerings offer
                       ON channels.offering = offer.id

                       LEFT JOIN accounts acc
                       ON channels.agent = acc.eth_addr

                       LEFT JOIN jobs job
                       ON channels.id = job.related_id
                 WHERE channels.agent NOT IN (SELECT eth_addr FROM accounts)
                       AND channels.id = job.related_id
		`
	queryFooter := `
		 GROUP BY channels.id, job.id, offer.setup_price, offer.unit_price,
                       offer.unit_type, offer.unit_name, offer.max_inactive_time_sec
		`

	conds, args := s.formatConditions(r, &getConf{
		Params: channelsGetParams,
	})

	query := queryHeader + s.filter(conds) + queryFooter

	resp, err := s.getClientChannelsItems(w, query, args)
	if err != nil {
		return
	}

	s.reply(logger, w, &resp)
}

// TODO(maxim) After the implementation of pagination, it is better to use this method for handleGetClientChannels.
// I specifically did not do the decomposition
/*
// handleGetClientChannels replies with all client channels or a channel by id
// available to the client.
func (s *Server) handleGetClientChannels(w http.ResponseWriter,
	r *http.Request) {
	resp := []*respGetClientChan{}

	conds, args := s.formatConditions(r, &getConf{
		Params: channelsGetParams,
	})

	tail := clientChannelsFilter + s.filter(conds)

	chs, err := s.db.SelectAllFrom(data.ChannelTable, tail, args...)
	if err != nil {
		s.logger.Warn("failed to select channels: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	for _, v := range chs {
		ch := v.(*data.Channel)

		var offer data.Offering
		var job data.Job

		if err := s.db.FindByPrimaryKeyTo(&offer,
			ch.Offering); err != nil {
			s.logger.Warn("failed to select offering: %v", err)
			s.replyUnexpectedErr(w)
			return
		}

		if err := s.db.FindOneTo(&job,
			"related_id", ch.ID); err != nil {
			s.logger.Warn("failed to select job: %v", err)
			s.replyUnexpectedErr(w)
			return
		}

		sessSlice, err := s.db.FindAllFrom(data.SessionTable,
			"channel", ch.ID)
		if err != nil {
			s.logger.Warn("failed to select channel: %v", err)
			s.replyUnexpectedErr(w)
		}

		var sess []*data.Session

		for _, v := range sessSlice {
			sess = append(sess, v.(*data.Session))
		}

		result := new(respGetClientChan)
		result.ID = ch.ID
		result.Agent = ethAddrFromHex(ch.Agent)
		result.Client = ethAddrFromHex(ch.Client)
		result.Offering = ch.Offering
		result.Deposit = ch.TotalDeposit
		result.ChStat.ChannelStatus = ch.ChannelStatus
		result.ChStat.ServiceStatus = ch.ServiceStatus
		result.ChStat.LastChanged = pointer.ToString(
			singleTimeFormat(*ch.ServiceChangedTime))
		if offer.MaxInactiveTimeSec != nil {
			result.ChStat.MaxInactiveTime = *offer.MaxInactiveTimeSec
		}

		result.Job.ID = job.ID
		result.Job.Type = job.Type
		result.Job.Status = job.Status
		result.Job.CreatedAt = singleTimeFormat(job.CreatedAt)

		var usage uint64
		var cost = offer.SetupPrice

		if offer.UnitType == data.UnitScalar {
			for _, ses := range sess {
				usage += ses.UnitsUsed
			}
		} else if offer.UnitType == data.UnitSeconds {
			for _, ses := range sess {
				usage += ses.SecondsConsumed
			}
		}
		cost += usage * offer.UnitPrice

		deposit := (ch.TotalDeposit - offer.SetupPrice) /
			offer.UnitPrice

		result.Usage.Cost = cost
		result.Usage.Current = usage
		result.Usage.MaxUsage = deposit
		result.Usage.Unit = offer.UnitType

		resp = append(resp, result)
	}
	s.reply(w, &resp)
}
*/

// handleGetChannelStatus replies with channels status by id.
func (s *Server) handleGetChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	logger := s.logger.Add("method", "handleGetChannelStatus")

	channel := &data.Channel{}
	if !s.findTo(logger, w, channel, id) {
		return
	}
	s.replyStatus(logger, w, channel.ChannelStatus)
}

// handleGetClientChannelStatus replies with client channels status by id.
func (s *Server) handleGetClientChannelStatus(w http.ResponseWriter,
	r *http.Request, id string) {
	logger := s.logger.Add("method", "handleGetClientChannelStatus")

	channel := new(data.Channel)
	if !s.selectOneTo(logger, w, channel, clientStatusFilter, id) {
		return
	}

	offer := new(data.Offering)
	if !s.findTo(logger, w, offer, channel.Offering) {
		return
	}

	resp := new(chanStatusBlock)

	if offer.MaxInactiveTimeSec == nil {
		offer.MaxInactiveTimeSec = new(uint64)
	} else {
		resp.MaxInactiveTime = *offer.MaxInactiveTimeSec
	}

	if channel.ServiceChangedTime == nil {
		resp.LastChanged = new(string)
	} else {
		resp.LastChanged = pointer.ToString(
			singleTimeFormat(*channel.ServiceChangedTime))
	}
	resp.ChannelStatus = channel.ChannelStatus
	resp.ServiceStatus = channel.ServiceStatus
	resp.MaxInactiveTime = *offer.MaxInactiveTimeSec

	s.reply(logger, w, &resp)
}

func (s *Server) putChannelStatus(w http.ResponseWriter, r *http.Request,
	id string, agent bool) {
	logger := s.logger.Add("method", "putChannelStatus", "id", id,
		"agent", agent)

	payload := &ActionPayload{}
	if !s.parsePayload(logger, w, r, payload) {
		return
	}

	logger.Info(fmt.Sprintf("action ( %v )  request for channel with id:"+
		" %v received.", payload.Action, id))

	if agent {
		if !s.findTo(logger, w, &data.Channel{}, id) {
			return
		}
	} else {
		if !s.selectOneTo(
			logger, w, &data.Channel{}, clientStatusFilter, id) {
			return
		}
	}

	var err error

	switch payload.Action {
	case channelPause:
		_, err = s.pr.SuspendChannel(id, data.JobUser, agent)
	case channelResume:
		_, err = s.pr.ActivateChannel(id, data.JobUser, agent)
	case channelTerminate:
		_, err = s.pr.TerminateChannel(id, data.JobUser, agent)
	case clientChannelClose:
		if !agent {
			s.createPreUncooperativeCloseRequest(w, id)
			return
		}
		s.replyInvalidRequest(logger, w)
		return
	default:
		s.replyInvalidAction(logger, w)
		return
	}

	if err != nil {
		logger.Error(fmt.Sprintf("failed to add job: %v", err))
		s.replyUnexpectedErr(logger, w)
	}
}

func (s *Server) createPreUncooperativeCloseRequest(w http.ResponseWriter, id string) {
	logger := s.logger.Add("method", "createPreUncooperativeCloseRequest",
		"id", id)

	gasPriceSettings := &data.Setting{}
	if err := data.FindOneTo(s.db.Querier, gasPriceSettings,
		"key", data.DefaultGasPriceKey); err != nil {
		logger.Error(err.Error())
		s.replyUnexpectedErr(logger, w)
		return
	}

	val, err := strconv.ParseInt(gasPriceSettings.Value, 10, 64)
	if err != nil {
		logger.Error(
			fmt.Sprintf("failed to parse default gas price: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	publishData, err := json.Marshal(&data.JobPublishData{
		GasPrice: uint64(val),
	})
	if err != nil {
		s.logger.Error(
			fmt.Sprintf("failed to marshal job publish data: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	err = s.queue.Add(&data.Job{
		Type:        data.JobClientPreUncooperativeCloseRequest,
		CreatedBy:   data.JobUser,
		RelatedID:   id,
		RelatedType: data.JobChannel,
		Data:        publishData,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to add job %s: %v",
			data.JobClientPreUncooperativeCloseRequest, err))
		s.replyUnexpectedErr(logger, w)
		return
	}
}

func (s *Server) handlePutChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	s.putChannelStatus(w, r, id, true)
}

func (s *Server) handlePutClientChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	s.putChannelStatus(w, r, id, false)
}
