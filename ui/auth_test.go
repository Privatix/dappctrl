package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"

	"github.com/privatix/dappctrl/util"
)

func TestSetPassword(t *testing.T) {
	err := handler.SetPassword("")
	util.TestExpectResult(t, "SetPassword", ui.ErrEmptyPassword, err)

	password := "foo"

	err = handler.SetPassword(password)
	util.TestExpectResult(t, "SetPassword", nil, err)

	salt := new(data.Setting)
	data.FindInTestDB(t, db, salt, "key", data.SettingPasswordSalt)

	passwordHash := new(data.Setting)
	data.FindInTestDB(t, db, passwordHash, "key",
		data.SettingPasswordHash)
	defer data.DeleteFromTestDB(t, db, salt, passwordHash)

	if data.ValidatePassword(passwordHash.Value, password, salt.Value) != nil {
		t.Fatal("password not set")
	}
}

func TestUpdatePassword(t *testing.T) {
	password := "foo"
	salt := "salt"
	hash, err := data.HashPassword(password, salt)
	util.TestExpectResult(t, "data.HashPassword", nil, err)

	saltSetting := &data.Setting{
		Key:   data.SettingPasswordSalt,
		Value: salt,
		Name:  "salt",
	}
	passwordSetting := &data.Setting{
		Key:   data.SettingPasswordHash,
		Value: hash,
		Name:  "password",
	}
	data.InsertToTestDB(t, db, saltSetting, passwordSetting)
	defer data.DeleteFromTestDB(t, db, saltSetting, passwordSetting)

	err = handler.UpdatePassword("wrong-password", "bar")
	util.TestExpectResult(t, "UpdatePassword", ui.ErrAccessDenied, err)

	err = handler.UpdatePassword(password, "")
	util.TestExpectResult(t, "UpdatePassword", ui.ErrEmptyPassword, err)

	newPassword := "new-password"

	err = handler.UpdatePassword(password, newPassword)
	util.TestExpectResult(t, "UpdatePassword", nil, err)

	db.Reload(passwordSetting)
	db.Reload(saltSetting)
	if data.ValidatePassword(
		passwordSetting.Value, newPassword, saltSetting.Value) != nil {
		t.Fatal("password not updated")
	}
}
