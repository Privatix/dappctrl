package billing

import (
	"gopkg.in/reform.v1"
	"errors"
	"github.com/privatix/dappctrl/util"
	"time"
)

type Monitor struct {
	interval time.Duration
	db *reform.DB
	logger *util.Logger
}

// todo: docs
func NewMonitor(interval time.Duration, db *reform.DB, logger * util.Logger) (*Monitor, error) {
	if db == nil {
		return nil, errors.New("`db` is required")
	}

	return &Monitor{interval, db, logger}, nil
}

func (m *Monitor) Run() error {
	m.logger.Info("Billing monitor started")

	for {
		time.Sleep(m.interval)
		m.logger.Info("Billing checks round started")

		if err := m.callChecksAndReportErrorIfAny(
			m.VerifySecondsBasedChannels,
			m.VerifyUnitsBasedChannels,
			m.VerifyChannelsForInactivity,
			m.processSuspendedChannels);

			err != nil {
			return err
		}

		m.logger.Info("Billing checks round finished")
	}
}

// Method is public, so it is possible to call it before providing any service,
// for example, in prepaid schemes. Also, this method might be called from the Rest API,
// as it is listed in the billing requirements.
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
		  offer.unit_type = 'seconds' AND
		  
		  (
		    offer.setup_price + (ses.seconds_consumed * offer.unit_price) >= channels.total_deposit OR
		    ses.seconds_consumed >= offer.max_unit
		  )`

	return m.processEachChannel(query, m.suspendService)
}

// Method is public, so it is possible to call it before providing any service,
// for example, in prepaid schemes. Also, this method might be called from the Rest API,
// as it is listed in the billing requirements.
func (m *Monitor) VerifyUnitsBasedChannels() error {
	query := `
		SELECT channels.id::text
		FROM channels
		  LEFT JOIN sessions ses ON channels.id = ses.channel
		  LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
		  channels.service_status IN ('pending', 'active') AND
		  offer.unit_type = 'units' AND
		  
		  (
		    offer.setup_price + (ses.units_used * offer.unit_price) >= channels.total_deposit OR
		    ses.units_used >= offer.max_unit
		  )`

	return m.processEachChannel(query, m.suspendService);
}

// todo: docs
func (m *Monitor) VerifyBillingLags() error {
	// Checking billing lags.
	// All channels, that are not suspended and are not terminated,
	// but are suffering from the billing lags - must be suspended.
	query := `
		SELECT channels.id::text
		FROM channels
		  LEFT JOIN sessions ses ON channels.id = ses.channel
		  LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
		  channels.service_status IN ('pending', 'active') AND
		  (
		    (ses.units_used / offer.billing_interval -
		      (channels.receipt_balance - offer.setup_price) / offer.unit_price) > offer.max_billing_unit_lag
		  )`

	if err := m.processEachChannel(query, m.suspendService); err != nil {
		return err
	}

	// All channels, that are suspended, but now seems to be payed - must be continued.
	query = `
		SELECT channels.id::text
		FROM channels
		  LEFT JOIN sessions ses ON channels.id = ses.channel
		  LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
		  channels.service_status = 'suspended' AND
		  (
		    (ses.units_used / offer.billing_interval -
		     (channels.receipt_balance - offer.setup_price) / offer.unit_price) <= offer.max_billing_unit_lag  
		  )`

	return m.processEachChannel(query, m.continueService)
}

// Method is public, so it is possible to call it before providing any service,
// for example, in prepaid schemes. Also, this method might be called from the Rest API,
// as it is listed in the billing requirements.
func (m *Monitor) VerifyChannelsForInactivity() error {
	query := `
		SELECT channels.id::text
		FROM channels
		  LEFT JOIN sessions ses ON channels.id = ses.channel
		  LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
		  channels.service_status IN ('pending', 'active')
		GROUP BY channels.id, offer.max_inactive_time_sec
		HAVING max(ses.last_usage_time) + (offer.max_inactive_time_sec * INTERVAL '1 second') < now()`

	return m.processEachChannel(query, m.suspendService)
}

// Method is public, so it is possible to call it before providing any service,
// for example, in prepaid schemes. Also, this method might be called from the Rest API,
// as it is listed in the billing requirements.
func (m *Monitor) processSuspendedChannels() error {
	query := `
		SELECT channels.id::text
		FROM channels
		  LEFT JOIN offerings offer ON channels.offering = offer.id
		
		WHERE
		  channels.service_status = 'suspended' AND
		    channels.service_changed_time + (offer.max_suspended_time * INTERVAL '1 second') < now()`

	return m.processEachChannel(query, m.suspendService)
}



func (m *Monitor) suspendService(uuid string) error {
	// todo: [call the task]
	return nil
}

func (m *Monitor) terminateService(uuid string) error {
	// todo: [call the task]
	return nil
}

func (m *Monitor) continueService(uuid string) error {
	// todo: [call the task]
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

	channelUUID := ""
	for rows.Next() {
		rows.Scan(&channelUUID)
		processor(channelUUID)
	}

	return nil
}
