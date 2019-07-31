package ui_test

import (
	"bytes"
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/client/somc"
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

	if data.ValidatePassword(
		data.Base64String(passwordHash.Value),
		password, salt.Value) != nil {
		t.Fatal("password not set")
	}
}

func updateAccountsPKeys(t *testing.T,
	accounts []reform.Struct, privateKey *ecdsa.PrivateKey) {
	newEncryptedKey, err := data.EncryptedKey(
		privateKey, data.TestPassword)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range accounts {
		v.(*data.Account).PrivateKey = newEncryptedKey
		data.SaveToTestDB(t, db, v.(*data.Account))
	}
}

func checkAccountsPKeys(t *testing.T, accounts []reform.Struct,
	privateKey *ecdsa.PrivateKey, password string) {
	for _, v := range accounts {
		acc := v.(*data.Account)
		db.Reload(acc)

		_, err := data.ToPrivateKey(
			acc.PrivateKey, data.TestPassword)
		if err == nil {
			t.Fatal("private key must not be decrypted")
		}

		pkAfter, err := data.ToPrivateKey(
			acc.PrivateKey, password)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(privateKey.D.Bytes(), pkAfter.D.Bytes()) {
			t.Fatal("private keys are not equal")
		}
	}
}

func TestUpdatePassword(t *testing.T) {
	server := rpc.NewServer()
	pwdStorage := data.NewPWDStorage(data.ToPrivateKey)
	pwdStorage.Set(data.TestPassword)
	handler := ui.NewHandler(logger, db, nil, pwdStorage,
		data.EncryptedKey, data.RoleClient, nil,
		somc.NewTestClientBuilder(testSOMCClient), testToken)
	err := server.RegisterName("ui2", handler)
	if err != nil {
		t.Fatal(err)
	}

	client := rpc.DialInProc(server)
	defer client.Close()

	fxt, assertMatchErr := newTest(t, "UpdatePassword")
	defer fxt.close()

	accounts, err := db.SelectAllFrom(data.AccountTable, "")
	if err != nil {
		t.Fatal(err)
	}

	privateKey, _ := crypto.GenerateKey()

	updateAccountsPKeys(t, accounts, privateKey)

	assertMatchErr(handler.UpdatePassword(
		"wrong-password", "bar"), ui.ErrAccessDenied)

	assertMatchErr(handler.UpdatePassword(
		data.TestPassword, ""), ui.ErrEmptyPassword)

	newPassword := "new-password"

	oldToken := testToken.v

	assertMatchErr(handler.UpdatePassword(
		data.TestPassword, newPassword), nil)

	if oldToken == testToken.v {
		t.Fatalf("token must be reset on password update")
	}

	db.Reload(fxt.hash)
	db.Reload(fxt.salt)

	if data.ValidatePassword(
		data.Base64String(fxt.hash.Value),
		newPassword, fxt.salt.Value) != nil {
		t.Fatal("password not updated")
	}

	checkAccountsPKeys(t, accounts, privateKey, newPassword)
}

func TestGetToken(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetToken")
	defer fxt.close()

	_, err := handler.GetToken("wrong-password")
	assertMatchErr(ui.ErrAccessDenied, err)

	v, err := handler.GetToken(data.TestPassword)
	assertMatchErr(nil, err)
	if v == nil || testToken.v != *v {
		t.Fatalf("wanted token %v, got %v", testToken.v, v)
	}
}
