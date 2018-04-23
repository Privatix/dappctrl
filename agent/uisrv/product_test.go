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
	return sendProductPayload(t, "POST", payload)
}

func putProduct(t *testing.T, payload *data.Product) *http.Response {
	return sendProductPayload(t, "PUT", payload)
}

func TestPostProductSuccess(t *testing.T) {
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

func createProductTestData(t *testing.T) *productTestData {
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	prod := data.NewTestProduct()
	prod.OfferTplID = &tplOffer.ID
	prod.OfferAccessID = &tplAccess.ID
	insertItems(t, tplOffer, tplAccess, prod)
	return &productTestData{tplOffer, tplAccess, prod}
}

func TestPutProduct(t *testing.T) {
	defer cleanDB(t)
	testData := createProductTestData(t)
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

func getProducts(t *testing.T) *http.Response {
	return getResources(t, productsPath, nil)
}

func testGetProducts(t *testing.T, exp int) {
	res := getProducts(t)
	testGetResources(t, res, exp)
}

func TestGetProducts(t *testing.T) {
	defer cleanDB(t)
	// Get empty list.
	testGetProducts(t, 0)
	// Get all products.
	createProductTestData(t)
	testGetProducts(t, 1)
}
