package sess

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
	reform "gopkg.in/reform.v1"
)

func (h *Handler) checkProductPassword(
	logger log.Logger, id, password string) (*data.Product, error) {
	var prod data.Product
	if err := h.db.FindByPrimaryKeyTo(&prod, id); err != nil {
		logger.Warn("failed to find product: " + err.Error())
		return nil, ErrAccessDenied
	}

	if err := data.ValidatePassword(
		prod.Password, password, string(prod.Salt)); err != nil {
		logger.Warn("failed to validate product password: " +
			err.Error())
		return nil, ErrAccessDenied
	}

	return &prod, nil
}

func (h *Handler) findClientChannel(logger log.Logger, product *data.Product,
	clientKey string, logError bool) (*data.Channel, error) {
	if product.ClientIdent != data.ClientIdentByChannelID {
		logger.Fatal("unsupported client identification type")
	}

	var ch data.Channel
	if err := h.db.FindByPrimaryKeyTo(&ch, clientKey); err != nil {
		msg := "failed to find channel: " + err.Error()
		if logError {
			logger.Error(msg)
		} else {
			logger.Warn(msg)
		}
		return nil, ErrChannelNotFound
	}
	return &ch, nil
}

func (h *Handler) findCurrentSession(
	logger log.Logger, channel string) (*data.Session, error) {
	var sess data.Session
	if err := h.db.SelectOneTo(&sess, `
		WHERE channel = $1 AND stopped IS NULL
		ORDER BY started DESC
		LIMIT 1`, channel); err != nil {
		if err == reform.ErrNoRows {
			logger.Warn("failed to find current session: " + err.Error())
			return nil, ErrSessionNotFound
		}
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	return &sess, nil
}
