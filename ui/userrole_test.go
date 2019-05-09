package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
)

func TestGetUserRole(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetUserRole")
	defer fxt.close()

	res, err := handler.GetUserRole()
	assertErrEqual(nil, err)
	if *res != data.RoleAgent && *res != data.RoleClient {
		t.Fatal("wrong user role")
	}
}
