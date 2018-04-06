package billing

import (
	"errors"
	"github.com/privatix/dappctrl/util"
	"gopkg.in/reform.v1"
	"time"
)

// Billing monitor provides logic for checking channels for various cases,
// in which service(s) must be suspended/terminated/or unsuspended (continued).
// All conditions are checked on the DB level,
// so it is safe to call monitors methods from separate goroutines.
type Monitor struct {
	db     *reform.DB
	logger *util.Logger

	// Interval between next round checks.
	interval time.Duration

	// Used only in tests purposes.
	// By default, in production builds is set to nil.
	// If not nil - collects IDs of channels, that was selected for further processing.
	testsSelectedChannelsIDs []string
}

func NewMonitor(interval time.Duration, db *reform.DB, logger *util.Logger) (*Monitor, error) {
	if db == nil {
		return nil, errors.New("`db` is required")
	}

	return &Monitor{db, logger, interval, nil}, nil
}

func (m *Monitor) Run() error {
	m.logger.Info("Billing monitor started")

	for {
		if err := m.processRound(); err != nil {
			return err
		}

		time.Sleep(m.interval)
	}
}

func (m *Monitor) VerifySecondsBasedChannels() error {
	// Selects all channels, which
	// 1. used tokens >= deposit tokens
	// 2. total consumed seconds >= max offer units (seconds in this case)
	// Only checks channels, which corresponding offers are using seconds as billing basis.
	query := `
                SELECT channels.id::text
                FROM channels
                        LEFT JOIN sessions ses ON channels.id = ses.channel
                        LEFT JOIN offerings offer ON channels.offering = offer.id
          
                WHERE
                        channels.service_status IN ('pending', 'active') AND
                        offer.unit_type = 'seconds'
          
                GROUP BY channels.id, offer.setup_price, offer.unit_price, offer.max_unit
                HAVING
                        (
				offer.setup_price + 
				coalesce(sum(ses.seconds_consumed), 0) * offer.unit_price 
				>= channels.total_deposit
                        ) OR 
			(
				coalesce(sum(ses.seconds_consumed), 0) >= offer.max_unit
			)`

	return m.processEachChannel(query, m.terminateService)
}

func (m *Monitor) VerifyUnitsBasedChannels() error {
	query := `
		SELECT channels.id::text
                FROM channels
                        LEFT JOIN sessions ses ON channels.id = ses.channel
                        LEFT JOIN offerings offer ON channels.offering = offer.id
          
                WHERE
                        channels.service_status IN ('pending', 'active') AND
                        offer.unit_type = 'units'
          
                GROUP BY channels.id, offer.setup_price, offer.unit_price, offer.max_unit
                HAVING
                        (
				offer.setup_price + 
				coalesce(sum(ses.units_used), 0) * offer.unit_price 
				>= channels.total_deposit
                        ) OR 
			(
				coalesce(sum(ses.units_used), 0) >= offer.max_unit
			)`

	return m.processEachChannel(query, m.terminateService)
}
func (m *Monitor) VerifyBillingLags() error {
	// Checking billing lags.
	// All channels, that are not suspended and are not terminated,
	// but are suffering from the billing lags - must be suspended.
	query := `
		SELECT channels.id :: text
                FROM channels
                        LEFT JOIN sessions ses ON channels.id = ses.channel
                        LEFT JOIN offerings offer ON channels.offering = offer.id

                WHERE
                        channels.service_status IN ('pending', 'active')

                GROUP BY channels.id, offer.billing_interval, offer.setup_price, offer.unit_price, offer.max_billing_unit_lag
                HAVING
                        (coalesce(sum(ses.units_used), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price) > offer.max_billing_unit_lag OR
                        (coalesce(sum(ses.seconds_consumed), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price) > offer.max_billing_unit_lag;`

	return m.processEachChannel(query, m.unsuspendService)
}

func (m *Monitor) VerifySuspendedChannelsAndTryToUnsuspend() error {
	// All channels, that are suspended, but now seems to be payed - must be unsuspended.
	query := `
		SELECT channels.id :: text
                FROM channels
                        LEFT JOIN sessions ses ON channels.id = ses.channel
                        LEFT JOIN offerings offer ON channels.offering = offer.id

                WHERE
                        channels.service_status IN ('suspended')

                GROUP BY channels.id, offer.billing_interval, offer.setup_price, offer.unit_price, offer.max_billing_unit_lag
                HAVING
                        (coalesce(sum(ses.units_used), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price) <= offer.max_billing_unit_lag OR
                        (coalesce(sum(ses.seconds_consumed), 0) / offer.billing_interval - (channels.receipt_balance - offer.setup_price) / offer.unit_price) <= offer.max_billing_unit_lag;`

	return m.processEachChannel(query, m.unsuspendService)
}

func (m *Monitor) VerifyChannelsForInactivity() error {
	query := `
		SELECT channels.id::text
		FROM channels
			LEFT JOIN sessions ses ON channels.id = ses.channel
			LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
			channels.service_status IN ('pending', 'active', 'suspended')
		GROUP BY channels.id, offer.max_inactive_time_sec
		HAVING max(ses.last_usage_time) + (offer.max_inactive_time_sec * INTERVAL '1 second') < now()`

	return m.processEachChannel(query, m.terminateService)
}

func (m *Monitor) VerifySuspendedChannelsAndTryToTerminate() error {
	query := `
		SELECT channels.id::text
		FROM channels
			LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
			channels.service_status = 'suspended' AND
			channels.service_changed_time + (offer.max_suspended_time * INTERVAL '1 second') < now()`

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
	// todo: [call the task] [simon]
	return nil
}

func (m *Monitor) terminateService(uuid string) error {
	// todo: [call the task] [simon]
	return nil
}

func (m *Monitor) unsuspendService(uuid string) error {
	// todo: [call the task] [simon]
	return nil
}

func (m *Monitor) callChecksAndReportErrorIfAny(checks ...func() error) error {
	for _, method := range checks {
		err := method()
		if err != nil {
			m.logger.Error("Internal billing error occurred. Details: ", err.Error())
			return err
		}
	}

	return nil
}

func (m *Monitor) processEachChannel(query string, processor func(string) error) error {
	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}

	if m.testsSelectedChannelsIDs != nil {
		m.testsSelectedChannelsIDs = make([]string, 0)
	}

	for rows.Next() {
		channelUUID := ""
		rows.Scan(&channelUUID)
		processor(channelUUID)

		if m.testsSelectedChannelsIDs != nil {
			m.testsSelectedChannelsIDs = append(m.testsSelectedChannelsIDs, channelUUID)
		}
	}

	return nil
}
