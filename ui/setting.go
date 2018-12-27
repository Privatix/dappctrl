package ui

import (
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

const settingsCondition = "WHERE permissions > 0"

// PermissionsToString associates a setting permissions with a title.
var PermissionsToString = map[int]string{
	data.ReadOnly:  "readOnly",
	data.ReadWrite: "readWrite",
}

// SettingUI is setting information.
type SettingUI struct {
	Value       string `json:"value"`
	Permissions string `json:"permissions"`
}

// GetSettings returns settings.
func (h *Handler) GetSettings(tkn string) (map[string]SettingUI, error) {
	logger := h.logger.Add("method", "GetSettings")

	if !h.token.Check(tkn) {
		return nil, ErrAccessDenied
	}

	result := make(map[string]SettingUI)

	settings, err := h.selectAllFrom(
		logger, data.SettingTable, settingsCondition)
	if err != nil {
		return nil, err
	}

	for _, v := range settings {
		setting := *v.(*data.Setting)
		result[setting.Key] = SettingUI{setting.Value,
			PermissionsToString[setting.Permissions]}
	}

	return result, err
}

// UpdateSettings updates settings.
func (h *Handler) UpdateSettings(tkn string, items map[string]string) error {
	logger := h.logger.Add("method", "UpdateSettings")

	if !h.token.Check(tkn) {
		return ErrAccessDenied
	}

	err := h.db.InTransaction(func(tx *reform.TX) error {
		for k, v := range items {
			if err := h.validateSetting(logger, k, v); err != nil {
				logger.Add("key", k, "value", v).Error(err.Error())
				return err
			}
		}

		for k, v := range items {
			logger = logger.Add("key", k, "value", v)

			var settingFromDB data.Setting

			// gets setting from database
			err := tx.FindByPrimaryKeyTo(&settingFromDB, k)
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

func (h *Handler) validateSetting(logger log.Logger, k, v string) error {
	// Run validators here.
	return nil
}
