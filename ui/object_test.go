package ui_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type testObject struct {
	oType string
	hash  data.HexString
}

type objectByHashResult struct {
	Hash data.HexString `json:"hash"`
}

func TestGetObjectByHash(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetObjectByHash")
	defer fxt.close()

	assertResult := func(wanted *testObject,
		res json.RawMessage, err error) {
		assertMatchErr(nil, err)

		var object *objectByHashResult
		if err := json.Unmarshal(res, &object); err != nil {
			t.Fatal(err)
		}

		if object.Hash != wanted.hash {
			t.Fatalf("hashes not equal, wanted: %s, got %s",
				wanted.hash, object.Hash)
		}
	}

	genHash := func() data.HexString {
		return data.HexFromBytes([]byte(util.NewUUID())[:32])
	}

	fxt.Endpoint.Hash = genHash()

	tx := &data.EthTx{
		ID:          util.NewUUID(),
		Hash:        genHash(),
		AddrFrom:    fxt.Account.EthAddr,
		Status:      data.TxSent,
		GasPrice:    1,
		Gas:         1,
		TxRaw:       []byte("{}"),
		RelatedType: data.JobOffering,
		RelatedID:   fxt.Offering.ID,
		Issued:      time.Now(),
	}

	data.SaveToTestDB(t, db, fxt.Endpoint)
	data.InsertToTestDB(t, db, tx)
	defer data.DeleteFromTestDB(t, db, tx)

	testData := []*testObject{
		{ui.TypeTemplate, fxt.TemplateOffer.Hash},
		{ui.TypeOffering, fxt.Offering.Hash},
		{ui.TypeEndpoint, fxt.Endpoint.Hash},
		{ui.TypeEthTx, tx.Hash},
	}

	_, err := handler.GetObjectByHash("wrong-password",
		ui.TypeOffering, string(fxt.Offering.Hash))
	assertMatchErr(ui.ErrAccessDenied, err)

	for _, v := range testData {
		res, err := handler.GetObjectByHash(data.TestPassword,
			v.oType, string(v.hash))
		assertResult(v, res, err)
	}

}
