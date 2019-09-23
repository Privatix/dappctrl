package sess

import (
	"time"

	"github.com/AlekSi/pointer"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/util"
)

// ServiceReady signals that service is ready to recieve connections. For agents only.
func (h *Handler) ServiceReady(product, productPassword, clientKey string) error {
	logger := h.logger.Add("method", "ServiceReady", "product", product,
		"clientKey", clientKey)

	logger.Info("session service ready request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, true)
	if err != nil {
		return err
	}

	if ch.ServiceStatus == data.ServiceActivating {
		err := job.AddWithData(h.queue, nil,
			data.JobCompleteServiceTransition,
			data.JobChannel, ch.ID, data.JobSessionServer,
			data.ServiceActive)
		if err != nil {
			return err
		}
	}
	return nil
}

// AuthClient verifies password for a given client key.
func (h *Handler) AuthClient(product, productPassword,
	clientKey, clientPassword string) error {
	logger := h.logger.Add("method", "AuthClient",
		"product", product, "clientKey", clientKey)

	logger.Info("session auth request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, false)
	if err != nil {
		return err
	}

	if ch.ServiceStatus != data.ServiceActive && ch.ServiceStatus != data.ServiceActivating {
		return ErrNonActiveChannel
	}

	err = data.ValidatePassword(
		ch.Password, clientPassword, string(ch.Salt))
	if err != nil {
		logger.Warn("failed to validate client password: " +
			err.Error())
		return ErrBadClientPassword
	}

	return nil
}

// StartSession creates a new client session.
func (h *Handler) StartSession(product, productPassword,
	clientKey, ip string, port uint16) (*data.Offering, error) {
	logger := h.logger.Add("method", "StartSession", "product", product,
		"clientKey", clientKey, "ip", ip, "port", port)

	logger.Info("session start request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return nil, err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, true)
	if err != nil {
		return nil, err
	}

	var offer data.Offering
	if err := h.db.FindByPrimaryKeyTo(&offer, ch.Offering); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	now := time.Now()

	var ipPtr *string
	if len(ip) != 0 {
		ipPtr = pointer.ToString(ip)
	}

	var portPtr *uint16
	if port != 0 {
		portPtr = pointer.ToUint16(port)
	}

	err = h.db.InTransaction(func(tx *reform.TX) error {
		sess := data.Session{
			ID:            util.NewUUID(),
			Channel:       ch.ID,
			Started:       now,
			LastUsageTime: now,
			ClientIP:      ipPtr,
			ClientPort:    portPtr,
		}
		if err := tx.Insert(&sess); err != nil {
			return err
		}

		if ch.ServiceStatus == data.ServiceActivating {
			err := job.AddWithData(h.queue, tx,
				data.JobCompleteServiceTransition,
				data.JobChannel, ch.ID, data.JobSessionServer,
				data.ServiceActive)
			if err != nil && err != job.ErrDuplicatedJob {
				return err
			}
		}

		return nil
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return &offer, nil
}

// UpdateSession updates current client session.
func (h *Handler) UpdateSession(product, productPassword, clientKey string,
	units uint64) error {
	logger := h.logger.Add("method", "UpdateSession", "product", product,
		"clientKey", clientKey, "units", units)

	logger.Info("session update request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, true)
	if err != nil {
		return err
	}

	// Make the server adapter to kill the session for non-active channel.
	if prod.IsServer && ch.ServiceStatus != data.ServiceActive {
		logger.Warn("non-active channel")
		return ErrNonActiveChannel
	}

	sess, err := h.findCurrentSession(logger, ch.ID)
	if err != nil {
		return err
	}
	logger = logger.Add("session", sess)

	if units != 0 {
		// TODO: Use unit size instead of this hardcode.
		units /= 1024 * 1024

		switch prod.UsageRepType {
		case data.ProductUsageIncremental:
			sess.UnitsUsed += units
		case data.ProductUsageTotal:
			sess.UnitsUsed = units
		default:
			logger.Fatal("unsupported product usage")
		}
	}

	sess.LastUsageTime = time.Now()

	logger.Info("updating session")

	if err := h.db.Save(sess); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

// StopSession stops current client session.
func (h *Handler) StopSession(product, productPassword, clientKey string) error {
	logger := h.logger.Add("method", "StopSession", "product", product,
		"clientKey", clientKey)

	logger.Info("session stop request")

	prod, err := h.checkProductPassword(logger, product, productPassword)
	if err != nil {
		return err
	}

	ch, err := h.findClientChannel(logger, prod, clientKey, true)
	if err != nil {
		return err
	}

	if ch.ServiceStatus == data.ServiceTerminated {
		logger.Warn("already terminated channel")
		return nil
	}

	status := data.ServiceSuspended
	if ch.ServiceStatus == data.ServiceTerminating {
		status = data.ServiceTerminated
	}

	err = job.AddWithData(h.queue, nil,
		data.JobCompleteServiceTransition,
		data.JobChannel, ch.ID, data.JobSessionServer,
		status)
	if err != nil && err != job.ErrDuplicatedJob {
		logger.Error(err.Error())
		return err
	}

	sess, err := h.findCurrentSession(logger, ch.ID)
	if err != nil {
		logger.Warn("Session not found")
		return nil
	}
	logger = logger.Add("session", sess)

	sess.LastUsageTime = time.Now()
	sess.Stopped = pointer.ToTime(sess.LastUsageTime)

	logger.Info("stopping session")

	if err := h.db.Save(sess); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}
