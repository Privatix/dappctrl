package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/ui"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestGetTransactions(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetTransactions")
	defer fxt.close()
	testRelID := util.NewUUID()
	tx := &data.EthTx{
		ID:          util.NewUUID(),
		Status:      data.TxSent,
		GasPrice:    1,
		Gas:         1,
		TxRaw:       []byte("{}"),
		RelatedType: data.JobChannel,
		RelatedID:   testRelID,
	}
	data.InsertToTestDB(t, db, tx)
	defer data.DeleteFromTestDB(t, db, tx)

	_, err := handler.GetTransactions("wrong-password", "", "")
	assertErrEqual(ui.ErrAccessDenied, err)

	assertResult := func(res []data.EthTx, err error, exp int) {
		assertErrEqual(nil, err)
		if len(res) != exp {
			t.Fatalf("wanted %d transactions, got %d", exp, len(res))
		}
	}

	res, err := handler.GetTransactions(
		data.TestPassword, "", "")
	assertResult(res, err, 1)

	res, err = handler.GetTransactions(
		data.TestPassword, data.JobChannel, "")
	assertResult(res, err, 1)

	res, err = handler.GetTransactions(
		data.TestPassword, "", testRelID)
	assertResult(res, err, 1)

	res, err = handler.GetTransactions(
		data.TestPassword, data.JobChannel, testRelID)
	assertResult(res, err, 1)

	res, err = handler.GetTransactions(
		data.TestPassword, data.JobOffering, "")
	assertResult(res, err, 0)

	res, err = handler.GetTransactions(
		data.TestPassword, "", util.NewUUID())
	assertResult(res, err, 0)
}
