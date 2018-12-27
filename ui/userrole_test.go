package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
)

func TestGetUserRole(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetUserRole")
	defer fxt.close()

	_, err := handler.GetUserRole("wrong-token")
	assertErrEqual(ui.ErrAccessDenied, err)

	res, err := handler.GetUserRole(testToken.v)
	assertErrEqual(nil, err)
	if *res != data.RoleAgent && *res != data.RoleClient {
		t.Fatal("wrong user role")
	}
}
