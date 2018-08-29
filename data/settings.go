package data

import (
	"fmt"
	"strconv"

	"gopkg.in/reform.v1"
)

// Setting keys.
const (
	SettingEthChallengePeriod = "eth.challenge.period"
	SettingPasswordHash       = "system.password"
	SettingPasswordSalt       = "system.salt"
	SettingAppVersion         = "system.version.app"
	SettingDefaultGasPrice    = "eth.default.gasprice"
	// SettingIsAgent specifies user role. "true" - agent. "false" - client.
	SettingIsAgent = "user.isagent"
)

// ReadSetting reads value of a given setting.
func ReadSetting(db *reform.Querier, key string) (string, error) {
	var st Setting
	if err := FindByPrimaryKeyTo(db, &st, key); err != nil {
		return "", fmt.Errorf(
			"failed to find '%s' setting: %s", key, err)
	}
	return st.Value, nil
}

func newSettingParseError(key string, err error) error {
	return fmt.Errorf("failed to parse '%s' setting value: %s", key, err)
}

// ReadUintSetting reads value of a given uint setting.
func ReadUintSetting(db *reform.Querier, key string) (uint, error) {
	val, err := ReadSetting(db, key)
	if err != nil {
		return 0, err
	}

	val2, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, newSettingParseError(key, err)
	}

	return uint(val2), nil
}
