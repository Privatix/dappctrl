// +build !noagentuisrvtest

package uisrv

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	reform "gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
)

func insertItems(items ...reform.Struct) func() {
	for _, item := range items {
		testServer.db.Insert(item)
	}
	return func() {
		for i := range items {
			rec := items[len(items)-i-1].(reform.Record)
			testServer.db.Delete(rec)
		}
	}
}

func createTestChannel() (*data.Channel, func()) {
	agent := data.NewTestUser()
	client := data.NewTestUser()
	product := data.NewTestProduct()
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	offering := data.NewTestOffering(agent.ID, product.ID, tplOffer.ID)
	ch := data.NewTestChannel(
		agent,
		client,
		offering,
		0,
		1,
		data.ChannelActive)
	deleteItems := insertItems(
		agent,
		client,
		product,
		tplOffer,
		offering,
		ch)
	return ch, deleteItems
}

func sendPayload(method,
	path string,
	payload interface{},
	handler func(http.ResponseWriter, *http.Request),
) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	json.NewEncoder(body).Encode(payload)
	r := httptest.NewRequest(method, offeringsPath, body)
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func getResources(path string,
	params map[string]string,
	handler func(http.ResponseWriter, *http.Request),
) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", path, nil)
	r.Form = make(url.Values)
	for k, v := range params {
		r.Form.Add(k, v)
	}
	w := httptest.NewRecorder()
	handler(w, r)
	return w
}

func testGetResources(t *testing.T, res *httptest.ResponseRecorder, exp int) {
	if res.Code != http.StatusOK {
		t.Fatal("failed to get products: ", res.Code)
	}
	resData := []map[string]interface{}{}
	json.NewDecoder(res.Body).Decode(&resData)
	if exp != len(resData) {
		t.Fatalf("expected %d items, got: %d", exp, len(resData))
	}
}
