package ui_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/privatix/dappctrl/ui"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func TestIncome(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetIncome")
	defer fxt.close()
	assertResult := func(exp uint, act *uint, err error) {
		assertErrEqual(nil, err)
		if act == nil || exp != *act {
			t.Fatalf("wrong result, wanted: %v, got: %v", exp, *act)
		}
	}

	ch1 := *fxt.Channel
	ch1.ID = util.NewUUID()
	ch1.ReceiptBalance = 10

	product := *fxt.Product
	product.ID = util.NewUUID()

	offering := *fxt.Offering
	offering.Product = product.ID
	offering.Hash = data.HexFromBytes(
		crypto.Keccak256([]byte(util.NewUUID())))
	offering.ID = util.NewUUID()

	ch2 := *fxt.Channel
	ch2.Offering = offering.ID
	ch2.ID = util.NewUUID()
	ch2.ReceiptBalance = 20

	ch3 := *fxt.Channel
	ch3.Offering = offering.ID
	ch3.ID = util.NewUUID()
	ch3.ReceiptBalance = 30

	data.InsertToTestDB(t, fxt.DB, &ch1, &product, &offering, &ch2, &ch3)
	defer data.DeleteFromTestDB(t, fxt.DB, &ch3, &ch2, &offering, &product,
		&ch1)

	// Test offerings income.
	_, err := handler.GetOfferingIncome("wrong-token", ch1.Offering)
	assertErrEqual(ui.ErrAccessDenied, err)

	actual, err := handler.GetOfferingIncome(testToken.v, fxt.Offering.ID)
	expected := uint(ch1.ReceiptBalance + fxt.Channel.ReceiptBalance)
	assertResult(expected, actual, err)

	// Test products income.
	_, err = handler.GetProductIncome("wrong-token", ch1.Offering)
	assertErrEqual(ui.ErrAccessDenied, err)

	actual, err = handler.GetProductIncome(testToken.v, product.ID)
	expected = uint(ch2.ReceiptBalance + ch3.ReceiptBalance)
	assertResult(expected, actual, err)

	// Test total income.
	actual, err = handler.GetTotalIncome(testToken.v)
	expected = uint(ch1.ReceiptBalance + ch2.ReceiptBalance +
		ch3.ReceiptBalance)
	assertResult(expected, actual, err)
}
