// +build !noagentuisrvtest

package uisrv

import (
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
)

const (
	keyRO  = "fooRO"
	keyRW  = "fooRW"
	keyAD  = "fooAD"
	keyNew = "fooNew"

	testValue    = "bar"
	changedValue = "changed"
)

var (
	settings = []data.Setting{
		{
			Key:         keyRO,
			Value:       testValue,
			Permissions: data.ReadOnly,
			Description: nil,
			Name:        keyRO,
		},
		{
			Key:         keyRW,
			Value:       testValue,
			Permissions: data.ReadWrite,
			Description: nil,
			Name:        keyRW,
		},
		{
			Key:         keyAD,
			Value:       testValue,
			Permissions: data.AccessDenied,
			Description: nil,
			Name:        keyAD,
		},
		{
			Key:         keyNew,
			Value:       testValue,
			Permissions: data.AccessDenied,
			Description: nil,
			Name:        keyNew,
		},
	}
)

func TestGetSettings(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// get empty list.
	testGetSettings(t, 0, "")

	insertSettings(t)

	testGetSettings(t, 2, "")
	testGetSettings(t, 1, keyRO)
	testGetSettings(t, 1, keyRW)
	testGetSettings(t, 0, keyAD)
}

func TestUpdateSettingsSuccess(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	insertSettings(t)

	settings[0].Value = changedValue
	settings[1].Value = changedValue
	settings[2].Value = changedValue
	settings[3].Value = changedValue

	res := putSetting(t, settingPayload(settings))
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to put setting: ", res.StatusCode)
	}

	result := allSettings(t)

	validateSetting(result, keyRO, testValue)
	validateSetting(result, keyRW, changedValue)
	validateSetting(result, keyAD, testValue)

	// can add a setting with settings.permissions = 0,
	// but can not read or change it
	testGetSettings(t, 0, keyNew)
	validateSetting(result, keyNew, testValue)
}

func allSettings(t *testing.T) (result map[string]data.Setting) {
	result = make(map[string]data.Setting)

	all, err := testServer.db.SelectAllFrom(data.SettingTable,
		"ORDER BY key")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range all {
		setting, ok := v.(*data.Setting)
		if !ok {
			continue
		}
		result[setting.Key] = *setting
	}
	return result
}

func insertSettings(t *testing.T) {
	insertItems(t, &settings[0], &settings[1], &settings[2], &settings[3])
}

func getSettings(t *testing.T, key string) *http.Response {
	return getResources(t, settingsPath, map[string]string{"key": key})
}

func testGetSettings(t *testing.T, exp int, key string) {
	res := getSettings(t, key)
	testGetResources(t, res, exp)
}

func putSetting(t *testing.T, pld settingPayload) *http.Response {
	return sendPayload(t, "PUT", settingsPath, pld)
}

func validateSetting(settings map[string]data.Setting,
	key, value string) bool {
	setting, ok := settings[key]
	if !ok {
		return false
	}
	return setting.Value == value
}
