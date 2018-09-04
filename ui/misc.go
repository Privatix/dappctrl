package ui

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

const (
	activeOfferingCondition = `
		offer_status = 'register'
			AND status = 'msg_channel_published'
			AND NOT is_local
			AND current_supply > 0
			AND agent NOT IN (SELECT eth_addr FROM accounts)`
)

func (h *Handler) findByPrimaryKey(logger log.Logger,
	notFoundError error, record reform.Record, id string) error {
	if err := h.db.FindByPrimaryKeyTo(record, id); err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return notFoundError
		}
		return ErrInternal
	}
	return nil
}

func (h *Handler) checkPassword(logger log.Logger, password string) error {
	hash, err := data.ReadSetting(h.db.Querier, data.SettingPasswordHash)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	salt, err := data.ReadSetting(h.db.Querier, data.SettingPasswordSalt)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	err = data.ValidatePassword(hash, password, salt)
	if err != nil {
		logger.Error(err.Error())
		return ErrAccessDenied
	}

	return nil
}

func (h *Handler) findActiveOfferingByID(
	logger log.Logger, id string) (*data.Offering, error) {
	var offer data.Offering
	if err := h.db.SelectOneTo(&offer,
		"WHERE id = $1 AND "+activeOfferingCondition, id); err != nil {
		logger.Error(err.Error())
		if err == reform.ErrNoRows {
			return nil, ErrOfferingNotFound
		}
		return nil, ErrInternal
	}
	return &offer, nil
}
