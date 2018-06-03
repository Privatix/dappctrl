package uisrv

import (
	"log"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/privatix/dappctrl/data"
)

var channelsGetParams = []queryParam{
	{Name: "id", Field: "id"},
	{Name: "channelStatus", Field: "channel_status"},
	{Name: "serviceStatus", Field: "service_status"},
}

type chanStatusBlock struct {
	ServiceStatus   string  `json:"serviceStatus"`
	ChannelStatus   string  `json:"channelStatus"`
	LastChanged     *string `json:"lastChanged"`
	MaxInactiveTime *uint64 `json:"maxInactiveTime"`
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

type RespGetClientChan struct {
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

// func returns ethereum's address on string format from base 64 encoded string
// if the address is not valid, it returns an empty string
func ethAddrFromBase64(addr string) string {
	ethAddr, err := data.ToAddress(addr)
	if err != nil {
		ethAddr = common.Address{}
	}
	return ethAddr.String()
}

func (s *Server) getClientChannelsItems(w http.ResponseWriter, query string,
	args []interface{}) (resp []*RespGetClientChan, err error) {
	resp = []*RespGetClientChan{}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		s.logger.Warn("failed to select: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	defer rows.Close()

	for rows.Next() {
		item := new(RespGetClientChan)
		i := new(usage)

		if err = rows.Scan(&item.ID, &item.Agent, &item.Client,
			&item.Offering, &item.Deposit,
			&item.ChStat.ServiceStatus, &item.ChStat.ChannelStatus,
			&item.ChStat.LastChanged, &item.ChStat.MaxInactiveTime,
			&item.Job.ID, &item.Job.Type, &item.Job.Status,
			&item.Job.CreatedAt, &i.secUsage, &i.unitsUsage,
			&item.Usage.MaxUsage, &i.unitType, &item.Usage.Unit,
			&i.costSeconds, &i.costUnits); err != nil {
			s.logger.Warn("failed to scan rows: %v", err)
			s.replyUnexpectedErr(w)
			return
		}

		// processing fields that can have a value of nil
		if item.ChStat.MaxInactiveTime == nil {
			item.ChStat.MaxInactiveTime = new(uint64)
		}

		if item.ChStat.LastChanged == nil {
			item.ChStat.LastChanged = new(string)
		}

		// client ETH address conversion
		item.Client = ethAddrFromBase64(item.Client)
		item.Agent = ethAddrFromBase64(item.Agent)

		// choosing the right type of channel
		if i.unitType == data.UnitScalar {
			item.Usage.Cost = i.costUnits
			item.Usage.Current = i.unitsUsage
		} else if i.unitType == data.UnitSeconds {
			item.Usage.Cost = i.costSeconds
			item.Usage.Current = i.secUsage
		}

		resp = append(resp, item)
	}
	if err = rows.Err(); err != nil {
		s.logger.Warn("failed to rows iteration: %v", err)
		s.replyUnexpectedErr(w)
		return
	}
	return resp, nil
}

// handleGetChannels replies with all channels or a channel by id
// available to the client.
func (s *Server) handleGetClientChannels(w http.ResponseWriter,
	r *http.Request) {
	// Result 20 fields: id, agent, client, offering, Deposit, service_status, channel_status,
	// last_changed, max_inactive_time_sec, job_id, job_type, job_status, job_created_at,
	// sec_usage, units_usage, max_usage, unit_type, unit_name, cost_seconds, cost_units
	queryHeader := `
		SELECT channels.id, channels.agent, channels.client, channels.offering,
                       channels.total_deposit AS Deposit, channels.service_status,
                       channels.channel_status, channels.service_changed_time AS last_changed,
                       offer.max_inactive_time_sec, job.id AS job_id, job.type AS job_type,
                       job.status AS job_status, job.created_at AS job_created_at,
                       COALESCE(SUM(ses.seconds_consumed), 0) AS sec_usage,
                       COALESCE(SUM(ses.units_used), 0) AS units_usage,
                       COALESCE(((channels.total_deposit - offer.setup_price) / offer.unit_price), 0) AS max_usage,
                       offer.unit_type, offer.unit_name,
                       COALESCE(offer.setup_price + COALESCE(SUM(ses.seconds_consumed), 0) * offer.unit_price) AS cost_seconds,
                       COALESCE(offer.setup_price + coalesce(sum(ses.units_used), 0) * offer.unit_price) AS cost_units
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
	queryFutter := `
		 GROUP BY channels.id, job.id, offer.setup_price, offer.unit_price,
                       offer.unit_type, offer.unit_name, offer.max_inactive_time_sec
		`

	conds, args := s.formatConditions(r, &getConf{
		Params: channelsGetParams,
	})

	constraints := s.filter(conds)

	query := queryHeader + constraints + queryFutter

	resp, err := s.getClientChannelsItems(w, query, args)
	if err != nil {
		return
	}

	s.reply(w, &resp)
}

// handleGetChannelStatus replies with channels status by id.
func (s *Server) handleGetChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	channel := &data.Channel{}
	if !s.findTo(w, channel, id) {
		return
	}
	s.replyStatus(w, channel.ChannelStatus)
}

func (s *Server) handleGetClientChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
}

const (
	channelTerminate = "terminate"
	channelPause     = "pause"
	channelResume    = "resume"
)

func (s *Server) handlePutChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	payload := &ActionPayload{}
	if !s.parsePayload(w, r, payload) {
		return
	}

	s.logger.Info("action ( %v )  request for channel with id: %v recieved.", payload.Action, id)

	jobTypes := map[string]string{
		channelTerminate: data.JobAgentPreServiceTerminate,
		channelPause:     data.JobAgentPreServiceSuspend,
		channelResume:    data.JobAgentPreServiceUnsuspend,
	}

	jobType, ok := jobTypes[payload.Action]
	if !ok {
		s.replyInvalidAction(w)
		return
	}

	if !s.findTo(w, &data.Channel{}, id) {
		return
	}

	if err := s.queue.Add(&data.Job{
		Type:        jobType,
		RelatedType: data.JobChannel,
		RelatedID:   id,
		CreatedBy:   data.JobUser,
		Data:        []byte("{}"),
	}); err != nil {
		s.logger.Error("failed to add job %s: %v", jobType, err)
		s.replyUnexpectedErr(w)
	}
}

func (s *Server) handlePutClientChannelStatus(w http.ResponseWriter, r *http.Request, id string) {
	log.Printf("id: %s", id)
}
