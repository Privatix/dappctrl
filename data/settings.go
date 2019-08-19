package data

import (
	"fmt"
	"strconv"

	"gopkg.in/reform.v1"
)

// Setting keys.
const (
	SettingAppVersion                       = "system.version.app"
	SettingBlockLimit                       = "eth.event.blocklimit"
	SettingOfferingsFreshBlocks             = "eth.event.offeringsfreshblocks"
	SettingOfferingAutoPopUp                = "offering.autopopup"
	SettingLastProcessedBlock               = "eth.event.lastProcessedBlock"
	SettingLastBackSearchBlock              = "eth.event.lastBackSearchBlock"
	SettingClientMonitoringStartBlock       = "eth.event.clientmonitoringStartBlock"
	SettingMinConfirmations                 = "eth.min.confirmations"
	SettingPasswordHash                     = "system.password"
	SettingPasswordSalt                     = "system.salt"
	SettingsPeriodChallange                 = "psc.periods.challenge"
	SettingsPeriodPopUp                     = "psc.periods.popup"
	SettingsPeriodRemove                    = "psc.periods.remove"
	SettingGUI                              = "system.gui"
	SettingClientMinDeposit                 = "client.min.deposit"
	SettingClientAutoincreaseDepositPercent = "client.autoincrease.percent"
	SettingClientAutoincreaseDeposit        = "client.autoincrease.deposit"
	SettingRatingRankingSteps               = "rating.ranking.steps"
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

// ReadUint64Setting reads value of a given uint setting.
func ReadUint64Setting(db *reform.Querier, key string) (uint64, error) {
	val, err := ReadSetting(db, key)
	if err != nil {
		return 0, err
	}

	val2, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, newSettingParseError(key, err)
	}

	return uint64(val2), nil
}

// UpdateUint64Setting reads value of a given uint setting.
func UpdateUint64Setting(db *reform.Querier, key string, val uint64) error {
	_, err := db.Exec(`UPDATE settings SET value=$1 WHERE key=$2`, val, key)
	return err
}

// ReadBoolSetting reads value of a given bool setting.
func ReadBoolSetting(db *reform.Querier, key string) (bool, error) {
	val, err := ReadSetting(db, key)
	if err != nil {
		return false, err
	}

	val2, err := strconv.ParseBool(val)
	if err != nil {
		return false, newSettingParseError(key, err)
	}

	return bool(val2), nil
}
