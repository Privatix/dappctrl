package ui_test

import (
	"encoding/json"
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
	}

	testData := []*testObject{
		{"", data.JobChannel, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobOffering, util.NewUUID()},
		{fxt.Account.EthAddr, data.JobAccount, util.NewUUID()},
	}

	for _, v := range testData {
		tx := &data.EthTx{
			ID:          util.NewUUID(),
			AddrFrom:    v.addrFrom,
			Status:      data.TxSent,
			GasPrice:    1,
			Gas:         1,
			TxRaw:       []byte("{}"),
			RelatedType: v.relatedType,
			RelatedID:   v.relatedID,
			Issued:      time.Now(),
		}

		data.InsertToTestDB(t, db, tx)
		defer data.DeleteFromTestDB(t, db, tx)
	}

	checkData := []*checkObject{
		// Test pagination. (remember one tx is part of fixture)
		{"", "", 0, 1, 1, 4},
		{"", "", 1, 3, 3, 4},
		// Test by filters.
		{"", "", 0, 0, 4, 4},
		{data.JobChannel, "", 0, 0, 1, 1},
		{"", testData[0].relatedID, 0, 0, 1, 1},
		{data.JobChannel, testData[0].relatedID, 0, 0, 1, 1},
		{data.JobEndpoint, "", 0, 0, 0, 0},
		{"", util.NewUUID(), 0, 0, 0, 0},
	}

	_, err := handler.GetEthTransactions("wrong-token", "", "", 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.GetEthTransactions(testToken.v,
		ui.AccountAggregatedType, "", 0, 0)
	assertErrEqual(ui.ErrInternal, err)

	_, err = handler.GetEthTransactions(testToken.v,
		ui.AccountAggregatedType, util.NewUUID(), 0, 0)
	assertErrEqual(ui.ErrAccountNotFound, err)

	for _, v := range checkData {
		res, err := handler.GetEthTransactions(testToken.v,
			v.relType, v.relID, v.offset, v.limit)
		assertErrEqual(nil, err)
		if len(res.Items) != v.exp {
			t.Logf("%+v\n", v)
			t.Fatalf("wanted %d, got %d", v.exp, len(res.Items))
		}
		if res.TotalItems != v.total {
			t.Fatalf("wanted %d, got %d", v.total, res.TotalItems)
		}
	}

	// Test accountAggregated type.
	res, err := handler.GetEthTransactions(testToken.v,
		ui.AccountAggregatedType, fxt.Account.ID, 0, 0)
	assertErrEqual(nil, err)
	if len(res.Items) != 2 {
		t.Fatalf("wanted %d, got %d", 2, len(res.Items))
	}
	if res.TotalItems != 2 {
		t.Fatalf("wanted %d, got %d", 2, res.TotalItems)
	}

	for _, v := range res.Items {
		if v.AddrFrom != fxt.Account.EthAddr {
			t.Fatalf("wanted Ethereum address: %s, got: %s",
				fxt.Account.EthAddr, v.AddrFrom)
		}
	}

	// Test ordering.
	res, err = handler.GetEthTransactions(testToken.v,
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

func TestIncreaseTxGasPrice(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "IncreaseTxGasPrice")
	defer fxt.close()

	err := handler.IncreaseTxGasPrice("wrong-token", "", 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	err = handler.IncreaseTxGasPrice(testToken.v, util.NewUUID(), 0)
	assertErrEqual(ui.ErrTxNotFound, err)

	err = handler.IncreaseTxGasPrice(testToken.v, fxt.EthTx.ID, fxt.EthTx.GasPrice-1)
	assertErrEqual(ui.ErrTxIsUnderpriced, err)

	j := new(data.Job)
	setTestJobQueueToExpectJobAdd(t, j)
	newGasPrice := fxt.EthTx.GasPrice + 1
	err = handler.IncreaseTxGasPrice(testToken.v, fxt.EthTx.ID, newGasPrice)
	assertErrEqual(nil, err)

	if j == nil || j.Type != data.JobIncreaseTxGasPrice ||
		j.RelatedType != data.JobTransaction || j.RelatedID != fxt.EthTx.ID {
		t.Fatalf("unexpected job: %v", j)
	}

	jdata := new(data.JobPublishData)
	json.Unmarshal(j.Data, jdata)
	if jdata.GasPrice != newGasPrice {
		t.Fatalf("wanted new gas price: %v, got: %v", newGasPrice, jdata.GasPrice)
	}
}
