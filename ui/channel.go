package ui

import (
	"fmt"
	"strings"

	"github.com/AlekSi/pointer"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// Actions that change a status of a channel.
const (
	ChannelTerminateAction = "terminate"
	ChannelPauseAction     = "pause"
	ChannelResumeAction    = "resume"
	ChannelCloseAction     = "close" // Client only.
)

var (
	clientChannelsCondition = `channels.agent NOT IN (SELECT eth_addr FROM accounts)`
	agentChannelsCondition  = `channels.agent IN (SELECT eth_addr FROM accounts)`
)

// GetAgentChannelsResult is result of GetAgentChannels method.
type GetAgentChannelsResult struct {
	Items      []data.Channel `json:"items"`
	TotalItems int            `json:"totalItems"`
}

// GetClientChannelsResult is result of GetClientChannels method.
type GetClientChannelsResult struct {
	Items      []ClientChannelInfo `json:"items"`
	TotalItems int                 `json:"totalItems"`
}

// ClientChannelInfo is information of client channel.
type ClientChannelInfo struct {
	ID           string         `json:"id"`
	Agent        string         `json:"agent"`
	Client       string         `json:"client"`
	Offering     string         `json:"offering"`
	OfferingHash data.HexString `json:"offeringHash"`
	Deposit      uint64         `json:"deposit"`

	ChStat chanStatusBlock `json:"channelStatus"`
	Job    jobBlock        `json:"job"`
	Usage  usageBlock      `json:"usage"`
}

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

// TopUpChannel initiates JobClientPreChannelTopUp job.
func (h *Handler) TopUpChannel(password, channel string, gasPrice uint64) error {
	logger := h.logger.Add("method", "TopUpChannel",
		"channel", channel, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	ch := &data.Channel{}
	if err := h.findByPrimaryKey(logger,
		ErrChannelNotFound, ch, channel); err != nil {
		return err
	}

	jdata, err := h.jobPublishData(logger, gasPrice)
	if err != nil {
		return err
	}

	return job.AddWithData(h.queue, nil, data.JobClientPreChannelTopUp,
		data.JobChannel, ch.ID, data.JobUser, jdata)
}

// ChangeChannelStatus updates channel state.
func (h *Handler) ChangeChannelStatus(password, channel, action string) error {
	logger := h.logger.Add("method", "ChangeChannelStatus",
		"channel", channel, "action", action, "userRole", h.userRole)

	if err := h.checkPassword(logger, password); err != nil {
		return err
	}

	condition := fmt.Sprintf("WHERE id = %s ", h.db.Placeholder(1))

	isAgent := h.userRole == data.RoleAgent

	if !isAgent {
		condition = fmt.Sprintf("%s AND %s",
			condition, clientChannelsCondition)
	}

	items, err := h.selectAllFrom(
		logger, data.ChannelTable, condition, channel)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	if len(items) != 1 {
		logger.Error(ErrChannelNotFound.Error())
		return ErrChannelNotFound
	}

	switch action {
	case ChannelPauseAction:
		_, err = h.processor.SuspendChannel(
			channel, data.JobUser, isAgent)
	case ChannelResumeAction:
		_, err = h.processor.ActivateChannel(
			channel, data.JobUser, isAgent)
	case ChannelTerminateAction:
		_, err = h.processor.TerminateChannel(
			channel, data.JobUser, isAgent)
	case ChannelCloseAction:
		if isAgent {
			logger.Error(ErrNotAllowedForAgent.Error())
			return ErrNotAllowedForAgent
		}
		if err := h.createPreUncooperativeCloseRequest(
			channel, logger); err != nil {
			return ErrInternal
		}
	default:
		logger.Warn(ErrBadAction.Error())
		return ErrBadAction
	}

	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// GetAgentChannels gets channels for agent.
func (h *Handler) GetAgentChannels(password string,
	channelStatus, serviceStatus []string,
	offset, limit uint) (*GetAgentChannelsResult, error) {
	logger := h.logger.Add("method", "GetAgentChannels",
		"channelStatus", channelStatus, "serviceStatus", serviceStatus)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	channels, total, err := h.getChannels(
		logger, channelStatus, serviceStatus,
		agentChannelsCondition, offset, limit)
	if err != nil {
		return nil, err
	}

	return &GetAgentChannelsResult{channels, total}, err
}

// GetClientChannels gets client channel information.
func (h *Handler) GetClientChannels(password, channelStatus,
	serviceStatus string, offset,
	limit uint) (*GetClientChannelsResult, error) {
	logger := h.logger.Add("method", "GetClientChannels",
		"channelStatus", channelStatus, "serviceStatus", serviceStatus)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}
	// TODO(maxim): Replace channelStatus, serviceStatus arguments types from string to []string.
	var chStatuses []string
	if channelStatus != "" {
		chStatuses = append(chStatuses, channelStatus)
	}
	var serStatus []string
	if serviceStatus != "" {
		serStatus = append(serStatus, serviceStatus)
	}

	chs, total, err := h.getChannels(logger, chStatuses, serStatus,
		clientChannelsCondition, offset, limit)
	if err != nil {
		return nil, err
	}

	items := make([]ClientChannelInfo, 0)
	for _, channel := range chs {
		result, err := h.createClientChannelResult(logger, &channel)
		if err != nil {
			return nil, err
		}

		items = append(items, *result)
	}

	return &GetClientChannelsResult{items, total}, nil
}

func (h *Handler) createPreUncooperativeCloseRequest(
	channel string, logger log.Logger) error {
	jobData, err := h.jobPublishData(logger, 0)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	err = job.AddWithData(h.queue, nil,
		data.JobClientPreUncooperativeCloseRequest, data.JobChannel,
		channel, data.JobUser, jobData)
	if err != nil {
		logger.Error(err.Error())
	}

	return err
}

func createChanStatusBlock(channel *data.Channel,
	offering *data.Offering) (result chanStatusBlock) {
	result.ChannelStatus = channel.ChannelStatus
	result.ServiceStatus = channel.ServiceStatus
	if channel.ServiceChangedTime != nil {
		result.LastChanged = pointer.ToString(
			util.SingleTimeFormat(*channel.ServiceChangedTime))
	}
	result.MaxInactiveTime = offering.MaxInactiveTimeSec

	return result
}

func createJobBlock(job2 *data.Job) (result jobBlock) {
	result.ID = job2.ID
	result.Type = job2.Type
	result.Status = job2.Status
	result.CreatedAt = util.SingleTimeFormat(job2.CreatedAt)

	return result
}

func createUsageBlock(channel *data.Channel, offering *data.Offering,
	sessions []*data.Session) (result usageBlock) {
	var usage uint64
	var cost = offering.SetupPrice

	if offering.UnitType == data.UnitScalar {
		for _, ses := range sessions {
			usage += ses.UnitsUsed
		}
	} else if offering.UnitType == data.UnitSeconds {
		for _, ses := range sessions {
			usage += ses.SecondsConsumed
		}
	}
	cost += usage * offering.UnitPrice

	deposit := (channel.TotalDeposit - offering.SetupPrice) /
		offering.UnitPrice

	result.Cost = cost
	result.Current = usage
	result.MaxUsage = deposit
	result.Unit = offering.UnitType

	return result
}

func (h *Handler) createClientChannelResult(logger log.Logger,
	channel *data.Channel) (result *ClientChannelInfo, err error) {
	result = &ClientChannelInfo{}

	offering := &data.Offering{}
	err = h.findByPrimaryKey(logger, ErrOfferingNotFound,
		offering, channel.Offering)
	if err != nil {
		return nil, err
	}

	jobCondition := fmt.Sprintf(`WHERE related_id = %s
					AND status IN ('%s', '%s')`,
		h.db.Placeholder(1),
		data.JobFailed, data.JobDone)

	tail := fmt.Sprintf(
		`%s AND created_at = (SELECT MAX(created_at) FROM jobs %s)`,
		jobCondition, jobCondition)

	item, err := h.selectOneFrom(h.logger, data.JobTable, ErrJobNotFound,
		tail, channel.ID)
	if err != nil {
		return nil, err
	}

	job2 := item.(*data.Job)

	logger = logger.Add("job", job2, "offering",
		offering, "channel", channel)

	sess, err := h.findAllFrom(logger, data.SessionTable,
		"channel", channel.ID)
	if err != nil {
		return nil, err
	}

	var sessions []*data.Session
	for _, v := range sess {
		sessions = append(sessions, v.(*data.Session))
	}

	result.ID = channel.ID
	agent, err := data.HexToAddress(channel.Agent)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	result.Agent = agent.String()

	client, err := data.HexToAddress(channel.Client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	result.Client = client.String()

	result.Offering = channel.Offering
	result.OfferingHash = offering.Hash
	result.Deposit = channel.TotalDeposit
	result.ChStat = createChanStatusBlock(channel, offering)
	result.Job = createJobBlock(job2)
	result.Usage = createUsageBlock(channel, offering, sessions)

	return result, err
}

func (h *Handler) getChannelsConditions(channelStatuses,
	serviceStatuses []string) (tail string, args []interface{}) {
	var conditions []string

	ph := 1

	statusCondition := func(arg []string, name string) string {
		phs := h.db.Placeholders(ph, len(arg))
		phSlice := strings.Join(phs, ",")
		ph = ph + len(arg)
		return fmt.Sprintf("%s IN (%s)", name, phSlice)
	}

	processStatuses := func(arg []string, name string) {
		if len(arg) != 0 {
			condition := statusCondition(arg, name)
			conditions = append(conditions, condition)
			for _, status := range arg {
				args = append(args, status)
			}
		}
	}

	processStatuses(channelStatuses, "channel_status")
	processStatuses(serviceStatuses, "service_status")

	if len(conditions) > 0 {
		tail = strings.Join(conditions, " AND ")
	}

	return tail, args
}

func (h *Handler) getChannels(logger log.Logger, channelStatus,
	serviceStatus []string, specCondition string,
	offset, limit uint) ([]data.Channel, int, error) {
	conditions, args := h.getChannelsConditions(
		channelStatus, serviceStatus)

	var tail string

	if conditions == "" {
		tail = fmt.Sprintf("WHERE %s", specCondition)
	} else {
		tail = fmt.Sprintf("WHERE %s AND %s",
			conditions, specCondition)
	}

	count, err := h.numberOfObjects(
		logger, data.ChannelTable.Name(), tail, args)
	if err != nil {
		return nil, 0, err
	}

	offsetLimit := h.offsetLimit(offset, limit)

	tail = fmt.Sprintf("%s %s", tail, offsetLimit)

	result, err := h.selectAllFrom(
		logger, data.ChannelTable, tail, args...)
	if err != nil {
		return nil, 0, err
	}

	channels := make([]data.Channel, len(result))
	for i, item := range result {
		channels[i] = *item.(*data.Channel)
	}

	return channels, count, err
}
