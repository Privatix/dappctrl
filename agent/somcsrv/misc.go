package somcsrv

import (
	"database/sql"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

func (h *Handler) offeringByHash(
	logger log.Logger, hash data.HexString) (*data.Offering, error) {
	offering := new(data.Offering)
	err := h.findOneTo(logger, offering, ErrOfferingNotFound, "hash", hash)
	return offering, err
}

func (h *Handler) offeringByID(
	logger log.Logger, id string) (*data.Offering, error) {
	offering := new(data.Offering)
	err := h.findOneTo(logger, offering, ErrOfferingNotFound, "id", id)
	return offering, err
}

func (h *Handler) endpointByChannelID(
	logger log.Logger, id string) (*data.Endpoint, error) {
	endpoint := new(data.Endpoint)
	err := h.findOneTo(logger, endpoint, ErrEndpointNotFound, "channel", id)
	return endpoint, err
}

func (h *Handler) findOneTo(logger log.Logger,
	str reform.Struct, errNotFound error, column string, val interface{}) error {
	if err := h.db.FindOneTo(str, column, val); err != nil {
		if err == sql.ErrNoRows {
			return errNotFound
		}
		logger.Error(err.Error())
		return ErrInternal
	}
	return nil
}
