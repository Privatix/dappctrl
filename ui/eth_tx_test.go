package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestGetEthTransactions(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetEthTransactions")
	defer fxt.close()

	type object struct {
		addrFrom    string
		relatedType string
		relatedID   string
	}

	testData := []*object{
		{"", data.JobChannel, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobOffering, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobAccount, util.NewUUID()},
	}

	for k := range testData {
		tx := &data.EthTx{
			ID:          util.NewUUID(),
			AddrFrom:    testData[k].addrFrom,
			Status:      data.TxSent,
			GasPrice:    1,
			Gas:         1,
			TxRaw:       []byte("{}"),
			RelatedType: testData[k].relatedType,
			RelatedID:   testData[k].relatedID,
		}

		data.InsertToTestDB(t, db, tx)
		defer data.DeleteFromTestDB(t, db, tx)
	}

	_, err := handler.GetEthTransactions("wrong-password", "", "")
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.GetEthTransactions(
		data.TestPassword, ui.AccountAggregatedType, "")
	assertErrEqual(ui.ErrInternal, err)

	_, err = handler.GetEthTransactions(
		data.TestPassword, ui.AccountAggregatedType, util.NewUUID())
	assertErrEqual(ui.ErrAccountNotFound, err)

	assertResult := func(res []data.EthTx, err error, exp int) {
		assertErrEqual(nil, err)
		if len(res) != exp {
			t.Fatalf("wanted %d transactions, got %d",
				exp, len(res))
		}
	}

	res, err := handler.GetEthTransactions(
		data.TestPassword, "", "")
	assertResult(res, err, 3)

	res, err = handler.GetEthTransactions(
		data.TestPassword, data.JobChannel, "")
	assertResult(res, err, 1)

	res, err = handler.GetEthTransactions(
		data.TestPassword, "", testData[0].relatedID)
	assertResult(res, err, 1)

	res, err = handler.GetEthTransactions(
		data.TestPassword, data.JobChannel, testData[0].relatedID)
	assertResult(res, err, 1)

	res, err = handler.GetEthTransactions(
		data.TestPassword, data.JobEndpoint, "")
	assertResult(res, err, 0)

	res, err = handler.GetEthTransactions(
		data.TestPassword, "", util.NewUUID())
	assertResult(res, err, 0)

	res, err = handler.GetEthTransactions(
		data.TestPassword, ui.AccountAggregatedType, fxt.Account.ID)
	assertResult(res, err, 2)

	for k := range res {
		if res[k].AddrFrom != fxt.Account.EthAddr {
			t.Fatalf("wanted Ethereum address: %s, got: %s",
				fxt.Account.EthAddr, res[k].AddrFrom)
		}
	}
}
