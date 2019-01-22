package ui

import (
	"encoding/json"
	"fmt"

	"github.com/privatix/dappctrl/data"
)

// GetGUISettings returns gui settings.
func (h *Handler) GetGUISettings(tkn string) (map[string]interface{}, error) {
	logger := h.logger.Add("method", "GetGUISettings")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	retStr, err := data.ReadSetting(h.db.Querier, data.SettingGUI)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	ret := make(map[string]interface{})

	err = json.Unmarshal([]byte(retStr), &ret)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return ret, nil
}

// SetGUISettings sets gui settings.
func (h *Handler) SetGUISettings(tkn string, v map[string]interface{}) error {
	logger := h.logger.Add("method", "SetGUISettings")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return ErrAccessDenied
	}

	d, err := json.Marshal(&v)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	_, err = h.db.Exec(`
		UPDATE settings 
		   SET value = $1
		 WHERE key=$2`, string(d), data.SettingGUI)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to set gui settings: %v", err))
		return ErrInternal
	}

	return nil
}
