// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
)

func validProductPayload(tplOffer, tplAccess string) data.Product {
	prod := data.NewTestProduct()
	prod.OfferTplID = &tplOffer
	prod.OfferAccessID = &tplOffer
	return *prod
}

func sendProductPayload(t *testing.T, m string, pld *data.Product) *http.Response {
	return sendPayload(t, m, productsPath, pld)
}

func postProduct(t *testing.T, payload *data.Product) *http.Response {
	return sendProductPayload(t, http.MethodPost, payload)
}

func putProduct(t *testing.T, payload *data.Product) *http.Response {
	return sendProductPayload(t, http.MethodPut, payload)
}

func TestPostProductSuccess(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	insertItems(t, tplOffer, tplAccess)
	payload := validProductPayload(tplOffer.ID, tplAccess.ID)
	res := postProduct(t, &payload)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("failed to post product: %d", res.StatusCode)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	product := &data.Product{}
	if err := testServer.db.FindByPrimaryKeyTo(product, reply.ID); err != nil {
		t.Fatal("failed to get product: ", err)
	}
}

func TestPostProductValidation(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	insertItems(t, tplOffer, tplAccess)
	validPld := validProductPayload(tplOffer.ID, tplAccess.ID)

	noOfferingTemplate := validPld
	noOfferingTemplate.OfferTplID = nil

	noAccessTemplate := validPld
	noAccessTemplate.OfferAccessID = nil

	noUsageRepType := validPld
	noUsageRepType.UsageRepType = ""

	invalidUsageRepType := validPld
	invalidUsageRepType.UsageRepType = "invalid-value"

	for _, payload := range []data.Product{
		noOfferingTemplate,
		noAccessTemplate,
		noUsageRepType,
		invalidUsageRepType,
	} {
		res := postProduct(t, &payload)
		if res.StatusCode != http.StatusBadRequest {
			t.Error("failed validation: ", res.StatusCode)
		}
	}
}

type productTestData struct {
	TplOffer  *data.Template
	TplAccess *data.Template
	Product   *data.Product
}

func createProductTestData(t *testing.T, agent bool) *productTestData {
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	prod := data.NewTestProduct()
	prod.OfferTplID = &tplOffer.ID
	prod.OfferAccessID = &tplAccess.ID
	prod.IsServer = agent
	insertItems(t, tplOffer, tplAccess, prod)
	return &productTestData{tplOffer, tplAccess, prod}
}

func TestPutProduct(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	testData := createProductTestData(t, true)
	payload := validProductPayload(testData.TplOffer.ID, testData.TplAccess.ID)
	payload.ID = testData.Product.ID
	res := putProduct(t, &payload)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("failed to put product: %d", res.StatusCode)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	updatedProduct := &data.Product{}
	testServer.db.FindByPrimaryKeyTo(updatedProduct, reply.ID)
	if updatedProduct.ID != testData.Product.ID ||
		reflect.DeepEqual(updatedProduct, testData.Product) {
		t.Fatal("product has not changed")
	}
}

func getProducts(t *testing.T, agent bool) *http.Response {
	if agent {
		return getResources(t, productsPath, nil)
	}
	return getResources(t, clientProductsPath, nil)
}

func testGetProducts(t *testing.T, exp int, agent bool) {
	res := getProducts(t, agent)
	testGetResources(t, res, exp)
}

func TestGetProducts(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Get empty list.
	testGetProducts(t, 0, true)

	// Get all products for Agent.
	createProductTestData(t, true)
	testGetProducts(t, 1, true)

	// Get all products for Client.
	createProductTestData(t, false)
	testGetProducts(t, 1, false)
}
