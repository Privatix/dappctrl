// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func postTemplate(t *testing.T, tpl *data.Template) *http.Response {
	return sendPayload(t, http.MethodPost, templatePath, tpl)
}

func TestPostTemplateValidation(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

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
		res := postTemplate(t, testcase.Payload)
		if testcase.Code != res.StatusCode {
			t.Errorf("unexpected reply code: %d", res.StatusCode)
			t.Logf("%+v", *testcase.Payload)
		}
	}
}

func TestPostTemplateSuccess(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

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
		res := postTemplate(t, &payload)
		if res.StatusCode != http.StatusCreated {
			t.Errorf("failed to create, response: %d", res.StatusCode)
		}
		reply := &replyEntity{}
		json.NewDecoder(res.Body).Decode(reply)
		tpl := &data.Template{}
		if err := testServer.db.FindByPrimaryKeyTo(tpl, reply.ID); err != nil {
			t.Errorf("failed to retrieve template, got: %v", err)
		}
	}
}

func getTemplates(t *testing.T, tplType, id string) *http.Response {
	return getResources(t, templatePath,
		map[string]string{"type": tplType, "id": id})
}

func testGetTemplates(t *testing.T, tplType, id string, exp int) {
	res := getTemplates(t, tplType, id)
	testGetResources(t, res, exp)
}

func TestGetTemplate(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

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
	insertItems(t, records[0], records[1], records[2])

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
