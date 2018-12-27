package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
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

func insertSettings(t *testing.T) {
	data.InsertToTestDB(t, db, &settings[0], &settings[1],
		&settings[2], &settings[3])
}

func deleteSettings(t *testing.T) {
	data.DeleteFromTestDB(t, db, &settings[0], &settings[1],
		&settings[2], &settings[3])
}

func testGetSettings(t *testing.T, exp int, err error,
	checkFunc func(error, error)) {
	res, err2 := handler.GetSettings(testToken.v)
	checkFunc(err, err2)
	if res == nil {
		t.Fatal("a result is nil")
	}

	if len(res) != exp {
		t.Fatalf("expected %d items, got: %d (%s)",
			exp, len(res), util.Caller())
	}

	for _, v := range res {
		if v.Permissions != ui.PermissionsToString[data.ReadOnly] &&
			v.Permissions != ui.PermissionsToString[data.ReadWrite] {
			t.Fatalf("permitions %s not valid permitions",
				v.Permissions)
		}
		if v.Value == "" {
			t.Fatal("setting value is empty")
		}
	}
}

func allSettings(t *testing.T) (result map[string]data.Setting) {
	result = make(map[string]data.Setting)

	all, err := db.SelectAllFrom(data.SettingTable,
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

func validateSetting(settings map[string]data.Setting,
	key, value string) bool {
	setting, ok := settings[key]
	if !ok {
		return false
	}
	return setting.Value == value
}

func TestGetSettings(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetSettings")
	defer fxt.close()

	_, err := handler.GetSettings("wrong-token")
	assertMatchErr(ui.ErrAccessDenied, err)

	testGetSettings(t, 0, nil, assertMatchErr)

	insertSettings(t)
	defer deleteSettings(t)

	testGetSettings(t, 2, nil, assertMatchErr)
}

func TestUpdateSettings(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "UpdateSettings")
	defer fxt.close()

	insertSettings(t)
	defer deleteSettings(t)

	update := make(map[string]string)

	for _, v := range settings {
		update[v.Key] = changedValue
	}

	err := handler.UpdateSettings("wrong-token", nil)
	assertMatchErr(ui.ErrAccessDenied, err)

	err = handler.UpdateSettings(testToken.v, update)
	assertMatchErr(nil, err)

	err = handler.UpdateSettings(
		testToken.v, map[string]string{"1": "2"})
	assertMatchErr(ui.ErrInternal, err)

	result := allSettings(t)

	validateSetting(result, keyRO, testValue)
	validateSetting(result, keyRW, changedValue)
	validateSetting(result, keyAD, testValue)

	// can add a setting with settings.permissions = data.AccessDenied,
	// but can not read or change it
	testGetSettings(t, 2, nil, assertMatchErr)
	validateSetting(result, keyNew, testValue)
}
