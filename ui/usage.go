package ui

import (
	"database/sql"

	"github.com/privatix/dappctrl/util/log"
)

// GetChannelUsage returns total units used for a given channel.
func (h *Handler) GetChannelUsage(password, channelID string) (*uint, error) {
	logger := h.logger.Add("method", "GetChannelUsage",
		"channelID", channelID)

	return h.usage(logger, password,
		`SELECT SUM(sessions.units_used)
		   FROM sessions
		  WHERE channel=$1`, channelID)
}

// GetOfferingUsage returns total units used for all channels
// with a given offering.
func (h *Handler) GetOfferingUsage(password, offeringID string) (*uint, error) {
	logger := h.logger.Add("method", "GetOfferingUsage",
		"offeringID", offeringID)

	return h.usage(logger, password,
		`SELECT SUM(sessions.units_used)
		   FROM channels
		   	JOIN sessions
			ON sessions.channel=channels.id
			   AND channels.offering=$1`,
		offeringID)
}

// GetProductUsage returns total units used in all channel
// of all offerings with given product.
func (h *Handler) GetProductUsage(password, productID string) (*uint, error) {
	logger := h.logger.Add("method", "GetProductUsage",
		"productID", productID)

	return h.usage(logger, password,
		`SELECT SUM(sessions.units_used)
		   FROM offerings
		   	JOIN channels
			ON channels.offering=offerings.id
			   AND offerings.product=$1
		   	JOIN sessions
		     	ON sessions.channel=channels.id`, productID)
}

func (h *Handler) usage(
	logger log.Logger, password, query, arg string) (*uint, error) {
	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	var queryRet sql.NullInt64
	err := h.db.QueryRow(query, arg).Scan(&queryRet)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	ret := uint(queryRet.Int64)
	return &ret, nil
}
