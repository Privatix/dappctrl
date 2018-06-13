// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestSetPasswordFirstTime(t *testing.T) {
	defer cleanDB(t)

	testPwd := "test-password"

	sendSetPasswordAndTestStatus(t, &passwordPayload{testPwd}, http.StatusCreated)

	testPasswordMatchesWithStored(t, testPwd)
}

func TestSetPasswordRepeat(t *testing.T) {
	defer cleanDB(t)

	insertItems(t, &data.Setting{Key: passwordKey})

	res := sendSetPasswordAndTestStatus(t,
		&passwordPayload{"test-password"}, http.StatusUnauthorized)

	testReplyErrorCode(t, res.Body, 0)
}

func TestSetPasswordAccountsExist(t *testing.T) {
	defer cleanDB(t)

	insertItems(t, data.NewTestAccount(testPassword))

	res := sendSetPasswordAndTestStatus(t,
		&passwordPayload{"test-password"}, http.StatusUnauthorized)

	testReplyErrorCode(t, res.Body, 1)
}

func TestSetPasswordOfWrongLen(t *testing.T) {
	defer cleanDB(t)

	sendSetPasswordAndTestStatus(t,
		&passwordPayload{"short"}, http.StatusBadRequest)

	p := &passwordPayload{"tooooooooooo-looooooooong"}
	sendSetPasswordAndTestStatus(t, p, http.StatusBadRequest)
}

func sendSetPasswordAndTestStatus(t *testing.T,
	p *passwordPayload, status int) *http.Response {
	return sendPayloadToAuthAndTestStatus(t, http.MethodPost, p, status)
}

func TestUpdatePassword(t *testing.T) {
	defer cleanDB(t)

	password := insertTestPassword(t)

	updatedPwd := password + "-updated"

	sendUpdatedPasswordAndTestStatus(t,
		&newPasswordPayload{password, updatedPwd}, http.StatusOK)

	testPasswordMatchesWithStored(t, updatedPwd)
}

func TestUpdatePasswordWrongCurrentPassword(t *testing.T) {
	defer cleanDB(t)

	password := insertTestPassword(t)

	updatedPwd := password + "-updated"

	sendUpdatedPasswordAndTestStatus(t,
		&newPasswordPayload{"wrong-pwd", updatedPwd},
		http.StatusUnauthorized)
}

func TestUpdatePasswordWrongLen(t *testing.T) {
	defer cleanDB(t)

	password := insertTestPassword(t)

	sendUpdatedPasswordAndTestStatus(t,
		&newPasswordPayload{password, "short"},
		http.StatusBadRequest)
}

func insertTestPassword(t *testing.T) string {
	password := "test-password"
	salt := util.NewUUID()
	hashed, _ := data.HashPassword(password, salt)
	insertItems(t, &data.Setting{Key: saltKey, Value: salt, Name: "SALT"},
		&data.Setting{Key: passwordKey, Value: string(hashed), Name: "PWD"})
	return password
}

func sendUpdatedPasswordAndTestStatus(t *testing.T,
	p *newPasswordPayload, status int) {
	sendPayloadToAuthAndTestStatus(t, "PUT", p, status)
}

func sendPayloadToAuthAndTestStatus(t *testing.T, method string,
	p interface{}, status int) *http.Response {
	res := sendPayload(t, method, authPath, p)
	if res.StatusCode != status {
		t.Fatalf("got: %d, wanted: %d", res.StatusCode, status)
	}
	return res
}

func testPasswordMatchesWithStored(t *testing.T, expected string) {
	salt := findSetting(t, saltKey)
	hashed := findSetting(t, passwordKey).Value

	err := data.ValidatePassword(hashed, expected, salt.Value)
	if err != nil {
		t.Fatal("wrong password stored")
	}
}

func findSetting(t *testing.T, key string) *data.Setting {
	rec := &data.Setting{}
	if err := testServer.db.FindByPrimaryKeyTo(rec, key); err != nil {
		t.Fatal("failed to get setting: ", err)
	}
	return rec
}

func TestBasicAuthMiddleware(t *testing.T) {
	defer cleanDB(t)

	var called bool
	var resRecorder *httptest.ResponseRecorder
	callHandlerAndTestStatus := func(pwd string, status int) {
		called = false
		handler := basicAuthMiddleware(testServer,
			func(http.ResponseWriter, *http.Request) { called = true })
		r := httptest.NewRequest("", "/", nil)
		if pwd != "" {
			r.SetBasicAuth("", pwd)
		}
		resRecorder = httptest.NewRecorder()
		handler(resRecorder, r)
		if resRecorder.Code != status {
			t.Fatalf("got: %d, wanted: %d",
				resRecorder.Code,
				http.StatusUnauthorized)
		}
	}

	// Test basic auth not provided.
	callHandlerAndTestStatus("", http.StatusUnauthorized)
	testReplyErrorCode(t, resRecorder.Body, 0)

	// Test password is not set up.
	callHandlerAndTestStatus("foo", http.StatusUnauthorized)
	testReplyErrorCode(t, resRecorder.Body, 1)

	salt := "salt"
	password := "test-password"
	passwordHash, _ := data.HashPassword(password, salt)
	insertItems(t, &data.Setting{Key: saltKey, Value: salt, Name: "SALT"},
		&data.Setting{Key: passwordKey, Value: string(passwordHash), Name: "PWD"})

	// Test wrong password.
	callHandlerAndTestStatus("wrong-pwd", http.StatusUnauthorized)
	testReplyErrorCode(t, resRecorder.Body, 0)

	// Test correct password.
	callHandlerAndTestStatus(password, http.StatusOK)
	if !called {
		t.Fatal("middleware unexpected return")
	}
}

func testReplyErrorCode(t *testing.T, res io.Reader, expected int) {
	errReply := &serverError{}
	json.NewDecoder(res).Decode(errReply)
	if errReply.Code != expected {
		t.Fatalf("got error code: %d, wanted: %d", errReply.Code, expected)
	}
}
