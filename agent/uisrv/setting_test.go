// +build !noagentuisrvtest

package uisrv

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
)

func getSettings() *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", settingsPath, nil)
	w := httptest.NewRecorder()
	testServer.handleSettings(w, r)
	return w
}

func testGetSettings(t *testing.T, exp int) {
	res := getSettings()
	testGetResources(t, res, exp)
}

func TestGetSettings(t *testing.T) {
	// get empty list.
	testGetSettings(t, 0)
	// get settings.
	setting := &data.Setting{
		Key:         "foo",
		Value:       "bar",
		Description: nil,
	}
	delete := insertItems(setting)
	defer delete()
	testGetSettings(t, 1)
}

func putSetting(pld settingPayload) *httptest.ResponseRecorder {
	return sendPayload("PUT", settingsPath, pld, testServer.handleSettings)
}

func TestUpdateSettingsSuccess(t *testing.T) {
	settings := []data.Setting{
		{
			Key:         "name1",
			Value:       "value1",
			Description: nil,
		},
		{
			Key:         "name2",
			Value:       "value2",
			Description: nil,
		},
	}
	deleteSettings := insertItems(&settings[0], &settings[1])
	defer deleteSettings()

	settings[0].Value = "changed"
	settings[1].Value = "changed"
	res := putSetting(settingPayload(settings))
	if res.Code != http.StatusOK {
		t.Fatal("failed to put setting: ", res.Code)
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
