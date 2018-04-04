// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func deleteAllTemplates() {
	testServer.db.DeleteFrom(data.TemplateTable, "")
}

func postTemplate(t *data.Template) *httptest.ResponseRecorder {
	return sendPayload("POST", templatePath, t, testServer.handleTempaltes)
}

func TestPostTemplateValidation(t *testing.T) {
	defer deleteAllTemplates()
	for _, testcase := range []struct {
		Payload *data.Template
		Code    int
	}{
		// Request without payload.
		{
			Payload: nil,
			Code:    http.StatusBadRequest,
		},
		// Wrong type.
		{
			Payload: &data.Template{
				Kind: "wrong-kind",
				Raw:  []byte("{}"),
			},
			Code: http.StatusBadRequest,
		},
		// Wrong format for src.
		{
			Payload: &data.Template{
				Kind: data.TemplateOffer,
				Raw:  []byte("not-json"),
			},
			Code: http.StatusBadRequest,
		},
	} {
		res := postTemplate(testcase.Payload)
		if testcase.Code != res.Code {
			t.Errorf("unexpected reply code: %d", res.Code)
			t.Logf("%+v", *testcase.Payload)
		}
	}
}

func TestPostTemplateSuccess(t *testing.T) {
	defer deleteAllTemplates()
	for _, payload := range []data.Template{
		{
			Kind: data.TemplateOffer,
			Raw:  []byte("{}"),
		},
		{
			Kind: data.TemplateAccess,
			Raw:  []byte("{}"),
		},
	} {
		res := postTemplate(&payload)
		if res.Code != http.StatusCreated {
			t.Errorf("failed to create, response: %d", res.Code)
		}
		reply := &replyEntity{}
		json.NewDecoder(res.Body).Decode(reply)
		tpl := &data.Template{}
		if err := testServer.db.FindByPrimaryKeyTo(tpl, reply.ID); err != nil {
			t.Errorf("failed to retrieve template, got: %v", err)
		}
	}
}

func getTemplates(tplType, id string) *httptest.ResponseRecorder {
	return getResources(templatePath,
		map[string]string{"type": tplType, "id": id},
		testServer.handleTempaltes)
}

func testGetTemplates(t *testing.T, tplType, id string, exp int) {
	res := getTemplates(tplType, id)
	testGetResources(t, res, exp)
}

func TestGetTemplate(t *testing.T) {
	// Get zerro templates.
	testGetTemplates(t, "", "", 0)

	// Prepare test data.
	records := []*data.Template{
		{
			ID:   util.NewUUID(),
			Kind: data.TemplateOffer,
			Raw:  []byte("{}"),
		},
		{
			ID:   util.NewUUID(),
			Kind: data.TemplateOffer,
			Raw:  []byte("{}"),
		},
		{
			ID:   util.NewUUID(),
			Kind: data.TemplateAccess,
			Raw:  []byte("{}"),
		},
	}
	insertItems(records[0], records[1], records[2])
	defer deleteAllTemplates()

	// Get all templates.
	testGetTemplates(t, "", "", len(records))
	// Get by id with a match.
	testGetTemplates(t, "", records[0].ID, 1)
	// Get by id, without matches.
	id := util.NewUUID()
	testGetTemplates(t, "", id, 0)
	// Get all by type.
	testGetTemplates(t, data.TemplateOffer, "", 2)
	testGetTemplates(t, data.TemplateAccess, "", 1)
	// Get by type and id with a match.
	id = records[1].ID
	testGetTemplates(t, data.TemplateOffer, id, 1)
	// Get by type and id without matches.
	testGetTemplates(t, data.TemplateAccess, id, 0)
}
