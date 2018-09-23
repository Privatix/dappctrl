package ui_test

import (
	"fmt"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
)

func TestGetLastBlockNumber(t *testing.T) {
	fxt, assertErrEquals := newTest(t, "GetLastBlockNumber")
	defer fxt.Close()
	assertBlockNumber := func(exp uint64, act *uint64, err error) {
		assertErrEquals(nil, err)
		if act == nil || exp != *act {
			t.Fatalf("wrong result, wanted: %v, got: %v", exp, act)
		}
	}

	_, err := handler.GetLastBlockNumber("wrong-password")
	assertErrEquals(ui.ErrAccessDenied, err)

	_, err = handler.GetLastBlockNumber(data.TestPassword)
	assertErrEquals(ui.ErrMinConfirmationsNotFound, err)

	// Insert min confirmations setting.
	minConfVal := uint64(100)
	setting := &data.Setting{
		Key:         data.SettingMinConfirmations,
		Value:       fmt.Sprint(minConfVal),
		Permissions: data.ReadWrite,
		Name:        "test min confirmations",
	}
	data.InsertToTestDB(t, fxt.DB, setting)
	defer data.DeleteFromTestDB(t, fxt.DB, setting)

	ret, err := handler.GetLastBlockNumber(data.TestPassword)
	assertBlockNumber(minConfVal, ret, err)

	// Populate db with test records.
	maxBlockStored := uint64(100)

	log1 := data.NewTestEthLog()
	log1.BlockNumber = maxBlockStored - 1
	log2 := data.NewTestEthLog()
	log2.BlockNumber = maxBlockStored
	data.InsertToTestDB(t, db, log1, log2)
	defer data.DeleteFromTestDB(t, db, log1, log2)

	ret, err = handler.GetLastBlockNumber(data.TestPassword)
	assertBlockNumber(maxBlockStored+minConfVal, ret, err)
}
