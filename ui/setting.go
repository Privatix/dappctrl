package ui

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

const settingsCondition = "WHERE permissions > 0"

// GetSettings returns settings.
func (h *Handler) GetSettings(
	password string) (map[string]string, error) {
	logger := h.logger.Add("method", "GetSettings")

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)

	settings, err := h.selectAllFrom(
		logger, data.SettingTable, settingsCondition)
	if err != nil {
		return nil, err
	}

	for _, v := range settings {
		setting := *v.(*data.Setting)
		result[setting.Name] = setting.Value
	}

	return result, err
}

// UpdateSettings updates settings.
func (h *Handler) UpdateSettings(password string,
	items map[string]string) error {
	logger := h.logger.Add("method", "UpdateSettings")

	err := h.checkPassword(logger, password)
	if err != nil {
		return err
	}

	err = h.db.InTransaction(func(tx *reform.TX) error {
		for k, v := range items {
			logger = logger.Add("key", k, "value", v)

			var settingFromDB data.Setting

			// gets setting from database
			err = tx.FindByPrimaryKeyTo(&settingFromDB, k)
			if err != nil {
				logger.Error(err.Error())
				return ErrInternal
			}

			// if settings.permissions != data.ReadWrite
			// then setting ignored
			if settingFromDB.Permissions != data.ReadWrite {
				continue
			}

			setting := settingFromDB
			setting.Value = v

			err = tx.Update(&setting)
			if err != nil {
				logger.Error(err.Error())
				return ErrInternal
			}
		}
		return nil
	})

	if err != nil {
		return h.catchError(logger, err)
	}
	return nil
}
