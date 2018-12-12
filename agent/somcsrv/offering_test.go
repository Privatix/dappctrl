package somcsrv_test

import (
	"encoding/json"
	"testing"

	"github.com/privatix/dappctrl/agent/somcsrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/util"
)

func TestGetOfferingMessage(t *testing.T) {
	// Test offering does not exist.
	_, err := handler.Offering("123")
	util.TestExpectResult(t, "Offering", somcsrv.ErrOfferingNotFound, err)

	// Test offering exists.
	fxt := data.NewTestFixture(t, db)
	defer fxt.Close()
	setOfferingRawMsg(t, fxt)
	rawMsg, err := handler.Offering(fxt.Offering.Hash)
	util.TestExpectResult(t, "Offering", nil, err)
	if fxt.Offering.RawMsg != *rawMsg {
		t.Fatalf("wanted: %s, got: %s", fxt.Offering.RawMsg, *rawMsg)
	}
}

func setOfferingRawMsg(t *testing.T, fxt *data.TestFixture) {
	offeringMsg := offer.OfferingMessage(fxt.Account, fxt.TemplateOffer,
		fxt.Offering)
	offeringMsgBytes, _ := json.Marshal(offeringMsg)
	key, _ := data.TestToPrivateKey(fxt.Account.PrivateKey, data.TestPassword)
	packed, _ := messages.PackWithSignature(offeringMsgBytes, key)
	fxt.Offering.RawMsg = data.FromBytes(packed)
	data.SaveToTestDB(t, db, fxt.Offering)
}
