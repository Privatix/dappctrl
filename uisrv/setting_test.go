// +build !noagentuisrvtest

package uisrv

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
)

func getSettings(t *testing.T) *http.Response {
	return getResources(t, settingsPath, nil)
}

func testGetSettings(t *testing.T, exp int) {
	res := getSettings(t)
	testGetResources(t, res, exp)
}

func TestGetSettings(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// get empty list.
	testGetSettings(t, 0)

	// get settings.
	setting := &data.Setting{
		Key:         "foo",
		Value:       "bar",
		Description: nil,
	}
	insertItems(t, setting)
	testGetSettings(t, 1)
}

func putSetting(t *testing.T, pld settingPayload) *http.Response {
	return sendPayload(t, "PUT", settingsPath, pld)
}

func TestUpdateSettingsSuccess(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	settings := []data.Setting{
		{
			Key:         "name1",
			Value:       "value1",
			Description: nil,
			Name:        "Name 1",
		},
		{
			Key:         "name2",
			Value:       "value2",
			Description: nil,
			Name:        "Name 2",
		},
	}
	insertItems(t, &settings[0], &settings[1])

	settings[0].Value = "changed"
	settings[1].Value = "changed"
	res := putSetting(t, settingPayload(settings))
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to put setting: ", res.StatusCode)
	}
	updatedSettings, err := testServer.db.SelectAllFrom(
		data.SettingTable,
		"order by key")
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(&settings[0], updatedSettings[0]) ||
		!reflect.DeepEqual(&settings[1], updatedSettings[1]) {
		t.Fatal("settings not updated")
	}
}
