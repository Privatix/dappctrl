package ui_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
)

func TestGetGUISettings(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetGUISettings")
	defer fxt.close()

	_, err := handler.GetGUISettings("wrong-token")
	assertErrEqual(ui.ErrAccessDenied, err)

	testdata := map[string]interface{}{
		"foo": false,
		"bar": 12.3,
		"bob": "robe",
	}

	d, _ := json.Marshal(&testdata)
	setting := &data.Setting{
		Key:         data.SettingGUI,
		Permissions: data.AccessDenied,
		Value:       string(d),
		Name:        data.SettingGUI,
	}
	data.InsertToTestDB(t, fxt.DB, setting)
	defer data.DeleteFromTestDB(t, fxt.DB, setting)

	ret, err := handler.GetGUISettings(testToken.v)
	assertErrEqual(nil, err)

	if !reflect.DeepEqual(testdata, ret) {
		t.Fatalf("wanted %v, got %v", testdata, ret)
	}
}

func TestSetGUISettings(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "SetGUISettings")
	defer fxt.close()

	err := handler.SetGUISettings("wrong-token", nil)
	assertErrEqual(ui.ErrAccessDenied, err)

	setting := &data.Setting{
		Key:         data.SettingGUI,
		Permissions: data.AccessDenied,
		Value:       "{}",
		Name:        data.SettingGUI,
	}
	data.InsertToTestDB(t, fxt.DB, setting)
	defer data.DeleteFromTestDB(t, fxt.DB, setting)

	testdata := map[string]interface{}{
		"foo": "bar",
	}
	err = handler.SetGUISettings(testToken.v, testdata)
	assertErrEqual(nil, err)

	raw, _ := data.ReadSetting(fxt.DB.Querier, data.SettingGUI)

	stored := make(map[string]interface{})
	json.Unmarshal([]byte(raw), &stored)

	if !reflect.DeepEqual(testdata, stored) {
		t.Fatalf("wanted %v, got %v", testdata, stored)
	}
}
