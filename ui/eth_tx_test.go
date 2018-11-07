package ui_test

import (
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func TestGetEthTransactions(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetEthTransactions")
	defer fxt.close()

	type testObject struct {
		addrFrom    data.HexString
		relatedType string
		relatedID   string
	}

	type checkObject struct {
		relType string
		relID   string
		offset  uint
		limit   uint
		exp     int
		total   int
	}

	assertResult := func(res *ui.GetEthTransactionsResult,
		err error, exp, total int) {
		assertErrEqual(nil, err)
		if len(res.Items) != exp {
			t.Fatalf("wanted %d, got %d", exp, len(res.Items))
		}
		if res.TotalItems != total {
			t.Fatalf("wanted %d, got %d", exp, res.TotalItems)
		}
	}

	testData := []*testObject{
		{"", data.JobChannel, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobOffering, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobAccount, util.NewUUID()},
	}

	checkData := []*checkObject{
		// Test pagination.
		{"", "", 0, 1, 1, 3},
		{"", "", 1, 3, 2, 3},
		// Test by filters.
		{"", "", 0, 0, 3, 3},
		{data.JobChannel, "", 0, 0, 1, 1},
		{"", testData[0].relatedID, 0, 0, 1, 1},
		{data.JobChannel, testData[0].relatedID, 0, 0, 1, 1},
		{data.JobEndpoint, "", 0, 0, 0, 0},
		{"", util.NewUUID(), 0, 0, 0, 0},
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
			Issued:      time.Now(),
		}

		data.InsertToTestDB(t, db, tx)
		defer data.DeleteFromTestDB(t, db, tx)
	}

	_, err := handler.GetEthTransactions("wrong-password", "", "", 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.GetEthTransactions(data.TestPassword,
		ui.AccountAggregatedType, "", 0, 0)
	assertErrEqual(ui.ErrInternal, err)

	_, err = handler.GetEthTransactions(data.TestPassword,
		ui.AccountAggregatedType, util.NewUUID(), 0, 0)
	assertErrEqual(ui.ErrAccountNotFound, err)

	for _, v := range checkData {
		res, err := handler.GetEthTransactions(data.TestPassword,
			v.relType, v.relID, v.offset, v.limit)
		assertResult(res, err, v.exp, v.total)
	}

	// Test accountAggregated type.
	res, err := handler.GetEthTransactions(data.TestPassword,
		ui.AccountAggregatedType, fxt.Account.ID, 0, 0)
	assertResult(res, err, 2, 2)

	for _, v := range res.Items {
		if v.AddrFrom != fxt.Account.EthAddr {
			t.Fatalf("wanted Ethereum address: %s, got: %s",
				fxt.Account.EthAddr, v.AddrFrom)
		}
	}

	// Test ordering.
	res, err = handler.GetEthTransactions(data.TestPassword,
		"", "", 0, 0)
	assertResult(res, err, 3, 3)

	first := res.Items[0].Issued

	for _, v := range res.Items {
		if v.Issued.After(first) {
			t.Fatalf("time %s after %s",
				v.Issued.String(), first.String())
		}
	}
}
