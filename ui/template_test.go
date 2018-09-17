package ui_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

func expectedTemplates(t *testing.T, resultsNumber int, tplType string,
	expected error, checkFunc func(error, error)) {
	res, err2 := handler.GetTemplates(data.TestPassword, tplType)
	checkFunc(expected, err2)

	if res == nil {
		return
	}

	if len(res) != resultsNumber {
		t.Fatalf("expected %d items, got: %d (%s)",
			resultsNumber, len(res), util.Caller())
	}

	for _, template := range res {
		if tplType != "" && template.Kind != tplType {
			t.Fatalf("invalid template type,"+
				" expected: %s, got %s",
				template.Kind, tplType)
		}
	}
}

func expectedTemplate(t *testing.T, id string, reference *data.Template,
	expected error, checkFunc func(error, error)) {
	res, err := handler.GetObject(
		data.TestPassword, ui.TypeTemplate, id)
	checkFunc(expected, err)

	if res == nil {
		return
	}

	var result *data.Template
	err = json.Unmarshal(res, &result)
	if err != nil {
		t.Fatal(err)
	}
	var resultMap, referenceMap map[string]string

	resultRaw := result.Raw
	referenceRaw := reference.Raw

	err = json.Unmarshal(resultRaw, &resultMap)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(referenceRaw, &referenceMap)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMap, referenceMap) ||
		result.ID != reference.ID ||
		result.Kind != reference.Kind ||
		result.Hash != reference.Hash {
		t.Fatal("invalid result template")
	}
}

func TestGetTemplates(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetTemplates")
	defer fxt.close()

	_, err := handler.GetTemplates("wrong-password", "")
	util.TestExpectResult(t, "GetTemplates", ui.ErrAccessDenied, err)

	// Get template by id.
	expectedTemplate(t,
		fxt.TemplateOffer.ID, fxt.TemplateOffer, nil, assertMatchErr)
	expectedTemplate(t,
		fxt.TemplateAccess.ID, fxt.TemplateAccess, nil, assertMatchErr)
	expectedTemplate(t,
		"", nil, ui.ErrObjectNotFound, assertMatchErr)
	expectedTemplate(t,
		util.NewUUID(), nil, ui.ErrObjectNotFound, assertMatchErr)

	// Get all templates.
	expectedTemplates(t, 2, "", nil, assertMatchErr)

	// Get by type.
	expectedTemplates(t, 1, fxt.TemplateOffer.Kind, nil, assertMatchErr)
	expectedTemplates(t, 1, fxt.TemplateAccess.Kind, nil, assertMatchErr)
	expectedTemplates(t, 1, "wrong-kind", ui.ErrInternal, assertMatchErr)
}

func TestCreateTemplate(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetTemplates")
	defer fxt.close()

	_, err := handler.CreateTemplate("wrong-password", fxt.TemplateOffer)
	assertMatchErr(ui.ErrAccessDenied, err)

	for _, template := range []*data.Template{
		{
			Kind: data.TemplateOffer,
			Raw: []byte(fmt.Sprintf(
				`{"fake" : "%s"}`, util.NewUUID())),
		},
		{
			Kind: data.TemplateAccess,
			Raw: []byte(fmt.Sprintf(
				`{"fake" : "%s"}`, util.NewUUID())),
		},
	} {
		res, err := handler.CreateTemplate(data.TestPassword, template)
		util.TestExpectResult(t, "CreateTemplate", nil, err)

		tpl := &data.Template{}
		err = db.FindByPrimaryKeyTo(tpl, res)
		if err != nil {
			t.Fatal(err)
		}
		data.DeleteFromTestDB(t, db, tpl)
	}
}
