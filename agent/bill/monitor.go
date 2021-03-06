package billing

import (
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util/log"
)

const (
	jobCreator = data.JobBillingChecker
)

// Config is a billing monitor configuration for agent.
type Config struct {
	Interval uint64 // In milliseconds.
}

// NewConfig creates a new billing monitor configuration for agent.
func NewConfig() *Config {
	return &Config{
		Interval: 5000,
	}
}

// Monitor provides logic for checking channels for various cases,
// in which service(s) must be suspended/terminated/or unsuspended (continued).
// All conditions are checked on the DB level,
// so it is safe to call monitors methods from separate goroutines.
type Monitor struct {
	db     *reform.DB
	logger log.Logger
	pr     *proc.Processor

	// Interval between next round checks.
	interval uint64
}

// NewMonitor creates new instance of billing monitor.
// 'interval' specifies how often channels checks must be performed.
func NewMonitor(interval uint64, db *reform.DB,
	logger log.Logger, pr *proc.Processor) (*Monitor, error) {
	if db == nil || logger == nil || pr == nil || interval == 0 {
		return nil, ErrInput
	}

	return &Monitor{
		db:       db,
		logger:   logger.Add("type", "agent/bill.Monitor"),
		pr:       pr,
		interval: interval,
	}, nil
}

// Run begins monitoring of channels.
// In case of error - doesn't restarts automatically.
func (m *Monitor) Run() error {
	m.logger.Info("Billing monitor started")

	tick := time.NewTicker(time.Duration(m.interval) * time.Millisecond)
	defer tick.Stop()
	for range tick.C {
		if err := m.processRound(); err != nil {
			return err
		}
	}
	return nil
}

/* TODO: uncomment when timebased billing will be implemented
// VerifySecondsBasedChannels checks all active seconds based channels
// for not using more units, than provided by quota and not exceeding
// over total deposit.
func (m *Monitor) VerifySecondsBasedChannels() error {
	// Selects all channels, which
	// 1. used tokens >= deposit tokens
	// 2. total consumed seconds >= max offer units (seconds in this case)
	// Only checks channels, which corresponding offers are using seconds
	// as billing basis.
	query := `
              SELECT channels.id::text
		FROM channels
                     LEFT JOIN sessions ses
		     ON channels.id = ses.channel

                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id
                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active')
                 AND channels.channel_status NOT IN ('pending')
                 AND offer.unit_type = 'seconds'
                 AND acc.in_use
               GROUP BY channels.id, offer.setup_price,
                     offer.unit_price, offer.max_unit
              HAVING offer.setup_price + COALESCE(SUM(ses.seconds_consumed), 0) * offer.unit_price >= channels.total_deposit
                  OR COALESCE(SUM(ses.seconds_consumed), 0) >= offer.max_unit;`

	return m.processEachChannel(query, m.terminateService)
}
*/

// VerifyUnitsBasedChannels checks all active units based channels
// for not using more units, than provided by quota
// and not exceeding over total deposit.
func (m *Monitor) VerifyUnitsBasedChannels() error {
	query := `
              SELECT channels.id::text
		FROM channels
                     LEFT JOIN sessions ses
                     ON channels.id = ses.channel

                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id
                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active')
                 AND channels.channel_status NOT IN ('pending')
                 AND offer.unit_type = 'units'
                 AND acc.in_use
               GROUP BY channels.id, offer.setup_price,
                     offer.unit_price, offer.max_unit
              HAVING offer.setup_price + coalesce(sum(ses.units_used), 0) * offer.unit_price >= channels.total_deposit
                  OR COALESCE(SUM(ses.units_used), 0) >= offer.max_unit;`

	logger := m.logger.Add("method", "VerifyUnitsBasedChannels")
	return m.processEachChannel(logger, query, m.terminateService)
}

// VerifyBillingLags checks all active channels for billing lags,
// and schedules suspending of those, who are suffering from billing lags.
func (m *Monitor) VerifyBillingLags() error {
	// Checking billing lags.
	// All channels, that are not suspended and are not terminated,
	// but are suffering from the billing lags - must be suspended.
	query := `
              SELECT channels.id :: text
		FROM channels
                     LEFT JOIN sessions ses
                     ON channels.id = ses.channel

                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id

                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active')
                     AND channels.channel_status NOT IN ('pending')
                     AND acc.in_use
               GROUP BY channels.id, offer.billing_interval,
                     offer.setup_price, offer.unit_price,
                     offer.max_billing_unit_lag
              HAVING COALESCE(SUM(ses.units_used), 0) /
	      offer.billing_interval - (channels.receipt_balance - offer.setup_price ) /
	      offer.unit_price > offer.max_billing_unit_lag;`
	logger := m.logger.Add("method", "VerifyBillingLags")
	return m.processEachChannel(logger, query, m.suspendService)
}

// VerifySuspendedChannelsAndTryToUnsuspend scans all supsended channels,
// and checks if all conditions are met to unsuspend them.
// Is so - schedules task for appropriate channel unsuspending.
func (m *Monitor) VerifySuspendedChannelsAndTryToUnsuspend() error {
	// All channels, that are suspended,
	// but now seems to be payed - must be unsuspended.
	query := `
              SELECT channels.id :: text
		FROM channels
                     LEFT JOIN sessions ses
                     ON channels.id = ses.channel

                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id

                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('suspended')
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
               GROUP BY channels.id, offer.billing_interval,
                     offer.setup_price, offer.unit_price,
                     offer.max_billing_unit_lag
              HAVING COALESCE(SUM(ses.units_used), 0) /
	      offer.billing_interval - (channels.receipt_balance - offer.setup_price) /
	      offer.unit_price <= offer.max_billing_unit_lag;`
	logger := m.logger.Add("method", "VerifySuspendedChannelsAndTryToUnsuspend")
	return m.processEachChannel(logger, query, m.unsuspendService)
}

// VerifyChannelsForInactivity scans all channels, that are not terminated,
// and terminates those of them, who are staying inactive too long.
func (m *Monitor) VerifyChannelsForInactivity() error {
	query := `
              SELECT channels.id::text
		FROM channels
                     LEFT JOIN sessions ses
                     ON channels.id = ses.channel
                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id
                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active', 'suspended')
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
               GROUP BY channels.id, offer.max_inactive_time_sec
              HAVING GREATEST(MAX(ses.last_usage_time), channels.service_changed_time) +
	      (offer.max_inactive_time_sec * INTERVAL '1 second') < now();`
	logger := m.logger.Add("method", "VerifyChannelsForInactivity")
	return m.processEachChannel(logger, query, m.terminateService)
}

// VerifySuspendedChannelsAndTryToTerminate scans all suspended channels,
// and terminates those of them, who are staying suspended too long.
func (m *Monitor) VerifySuspendedChannelsAndTryToTerminate() error {
	query := `
              SELECT channels.id::text
		FROM channels
                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id

                     INNER JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status = 'suspended'
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
                 AND channels.service_changed_time + (offer.max_suspended_time * INTERVAL '1 SECOND') < now();`
	logger := m.logger.Add("method", "VerifySuspendedChannelsAndTryToTerminate")
	return m.processEachChannel(logger, query, m.terminateService)
}

func (m *Monitor) processRound() error {
	return m.callChecksAndReportErrorIfAny(
		// TODO: uncomment when timebased billing will be implemented
		/*m.VerifySecondsBasedChannels,*/
		m.VerifyUnitsBasedChannels,
		m.VerifyChannelsForInactivity,
		m.VerifySuspendedChannelsAndTryToUnsuspend,
		m.VerifySuspendedChannelsAndTryToTerminate,
		m.VerifyBillingLags)
}

func (m *Monitor) suspendService(uuid string) error {
	logger := m.logger.Add("method", "suspendService", "channel", uuid)
	_, err := m.pr.SuspendChannel(uuid, jobCreator, true)
	if err != nil {
		logger.Info(err.Error())
	}
	return nil
}

func (m *Monitor) terminateService(uuid string) error {
	logger := m.logger.Add("method", "terminateService", "channel", uuid)
	_, err := m.pr.TerminateChannel(uuid, jobCreator, true)
	if err != nil {
		logger.Info(err.Error())
	}
	return nil
}

func (m *Monitor) unsuspendService(uuid string) error {
	logger := m.logger.Add("method", "unsuspendService", "channel", uuid)
	_, err := m.pr.ActivateChannel(uuid, jobCreator, true)
	if err != nil {
		logger.Info(err.Error())
	}
	return nil
}

func (m *Monitor) callChecksAndReportErrorIfAny(checks ...func() error) error {
	for _, method := range checks {
		err := method()
		if err != nil {
			m.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (m *Monitor) processEachChannel(logger log.Logger, query string,
	processor func(string) error) error {
	rows, err := m.db.Query(query)
	defer rows.Close()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	for rows.Next() {
		channelUUID := ""
		if err := rows.Scan(&channelUUID); err != nil {
			logger.Error(err.Error())
		}
		logger.Info("processing channel: " + channelUUID)
		if err := processor(channelUUID); err != nil {
			return err
		}
	}

	return nil
}
