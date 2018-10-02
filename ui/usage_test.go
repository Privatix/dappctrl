package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
)

func TestUsage(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "Usage")
	defer fxt.close()

	assertUsage := func(exp uint, act *uint, err error) {
		assertErrEqual(nil, err)
		if act == nil || exp != *act {
			t.Fatalf("wrong usage, wanted: %v, got: %v", exp, act)
		}
	}

	// Prepare 2 sessions with different channels for the same offering
	// and product.
	sess1 := data.NewTestSession(fxt.Channel.ID)
	sess1.UnitsUsed = 10

	channel2 := data.NewTestChannel(fxt.Account.EthAddr, fxt.User.EthAddr,
		fxt.Offering.ID, 0, 10, data.ChannelActive)

	sess2 := data.NewTestSession(channel2.ID)
	sess2.UnitsUsed = 20

	data.InsertToTestDB(t, fxt.DB, sess1, channel2, sess2)
	defer data.DeleteFromTestDB(t, fxt.DB, sess2, channel2, sess1)

	// Test GetChannelUsage.
	_, err := handler.GetChannelUsage("wrong-password", fxt.Channel.ID)
	assertErrEqual(ui.ErrAccessDenied, err)

	ret, err := handler.GetChannelUsage(data.TestPassword, fxt.Channel.ID)
	assertUsage(uint(sess1.UnitsUsed), ret, err)

	// Test GetOfferingUsage.
	_, err = handler.GetOfferingUsage("wrong-password", fxt.Offering.ID)
	assertErrEqual(ui.ErrAccessDenied, err)

	ret, err = handler.GetOfferingUsage(data.TestPassword, fxt.Offering.ID)
	assertUsage(uint(sess1.UnitsUsed+sess2.UnitsUsed), ret, err)

	// Test GetProductUsage.
	_, err = handler.GetProductUsage("wrong-password", fxt.Product.ID)
	assertErrEqual(ui.ErrAccessDenied, err)

	ret, err = handler.GetProductUsage(data.TestPassword, fxt.Product.ID)
	assertUsage(uint(sess1.UnitsUsed+sess2.UnitsUsed), ret, err)
}
