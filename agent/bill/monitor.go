package billing

import (
	"time"

	"github.com/pkg/errors"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
)

const (
	jobCreator = data.JobBillingChecker
)

// Billing monitor specific errors.
var (
	ErrInput = errors.New("one or more input parameters is wrong")
)

// Monitor provides logic for checking channels for various cases,
// in which service(s) must be suspended/terminated/or unsuspended (continued).
// All conditions are checked on the DB level,
// so it is safe to call monitors methods from separate goroutines.
type Monitor struct {
	db     *reform.DB
	logger *util.Logger
	pr     *proc.Processor

	// Interval between next round checks.
	interval time.Duration
}

// NewMonitor creates new instance of billing monitor.
// 'interval' specifies how often channels checks must be performed.
func NewMonitor(interval time.Duration, db *reform.DB,
	logger *util.Logger, pc *proc.Processor) (*Monitor, error) {
	if db == nil || logger == nil || pc == nil || interval <= 0 {
		return nil, ErrInput
	}

	return &Monitor{db, logger, pc, interval}, nil
}

// Run begins monitoring of channels.
// In case of error - doesn't restarts automatically.
func (m *Monitor) Run() error {
	m.logger.Info("Billing monitor started")

	for {
		if err := m.processRound(); err != nil {
			return err
		}

		time.Sleep(m.interval)
	}
}

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

                     LEFT JOIN accounts acc
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

                     LEFT JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active')
                 AND channels.channel_status NOT IN ('pending')
                 AND offer.unit_type = 'units'
                 AND acc.in_use
               GROUP BY channels.id, offer.setup_price,
                     offer.unit_price, offer.max_unit
              HAVING offer.setup_price + coalesce(sum(ses.units_used), 0) * offer.unit_price >= channels.total_deposit
                  OR COALESCE(SUM(ses.units_used), 0) >= offer.max_unit;`

	return m.processEachChannel(query, m.terminateService)
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

                     LEFT JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active')
                     AND channels.channel_status NOT IN ('pending')
                     AND acc.in_use
               GROUP BY channels.id, offer.billing_interval,
                     offer.setup_price, offer.unit_price,
                     offer.max_billing_unit_lag
              HAVING COALESCE(SUM(ses.units_used), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price ) / offer.unit_price > offer.max_billing_unit_lag
                  OR COALESCE(SUM(ses.seconds_consumed), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price > offer.max_billing_unit_lag;`

	return m.processEachChannel(query, m.suspendService)
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

                     LEFT JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('suspended')
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
               GROUP BY channels.id, offer.billing_interval,
                     offer.setup_price, offer.unit_price,
                     offer.max_billing_unit_lag
              HAVING COALESCE(SUM(ses.units_used), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price <= offer.max_billing_unit_lag
                  OR COALESCE(SUM(ses.seconds_consumed), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price <= offer.max_billing_unit_lag;`

	return m.processEachChannel(query, m.unsuspendService)
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

                     LEFT JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status IN ('pending', 'active', 'suspended')
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
               GROUP BY channels.id, offer.max_inactive_time_sec
              HAVING MAX(ses.last_usage_time) + (offer.max_inactive_time_sec * INTERVAL '1 second') < now();`

	return m.processEachChannel(query, m.terminateService)
}

// VerifySuspendedChannelsAndTryToTerminate scans all suspended channels,
// and terminates those of them, who are staying suspended too long.
func (m *Monitor) VerifySuspendedChannelsAndTryToTerminate() error {
	query := `
              SELECT channels.id::text
		FROM channels
                     LEFT JOIN offerings offer
                     ON channels.offering = offer.id

                     LEFT JOIN accounts acc
                     ON channels.agent = acc.eth_addr
               WHERE channels.service_status = 'suspended'
                 AND channels.channel_status NOT IN ('pending')
                 AND acc.in_use
                 AND channels.service_changed_time + (offer.max_suspended_time * INTERVAL '1 SECOND') < now();`

	return m.processEachChannel(query, m.terminateService)
}

func (m *Monitor) processRound() error {
	return m.callChecksAndReportErrorIfAny(
		m.VerifySecondsBasedChannels,
		m.VerifyUnitsBasedChannels,
		m.VerifyChannelsForInactivity,
		m.VerifySuspendedChannelsAndTryToUnsuspend,
		m.VerifySuspendedChannelsAndTryToTerminate)
}

func (m *Monitor) suspendService(uuid string) error {
	_, err := m.pr.SuspendChannel(uuid, jobCreator)
	return err
}

func (m *Monitor) terminateService(uuid string) error {
	_, err := m.pr.TerminateChannel(uuid, jobCreator)
	return err
}

func (m *Monitor) unsuspendService(uuid string) error {
	_, err := m.pr.ActivateChannel(uuid, jobCreator)
	return err
}

func (m *Monitor) callChecksAndReportErrorIfAny(checks ...func() error) error {
	for _, method := range checks {
		err := method()
		if err != nil {
			m.logger.Error("Internal billing error occurred."+
				" Details: %s", err.Error())
			return err
		}
	}

	return nil
}

func (m *Monitor) processEachChannel(query string,
	processor func(string) error) error {
	rows, err := m.db.Query(query)
	defer rows.Close()
	if err != nil {
		return err
	}

	for rows.Next() {
		channelUUID := ""
		rows.Scan(&channelUUID)
		if err := processor(channelUUID); err != nil {
			return err
		}
	}

	return nil
}
