// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

func validProductPayload(tplOffer, tplAccess string) data.Product {
	return data.Product{
		Name:          "test-product",
		OfferTplID:    &tplOffer,
		OfferAccessID: &tplOffer,
		UsageRepType:  data.ProductUsageIncremental,
	}
}

func sendProductPayload(m string, pld *data.Product) *httptest.ResponseRecorder {
	return sendPayload(m, productsPath, pld, testServer.handleProducts)
}

func postProduct(payload *data.Product) *httptest.ResponseRecorder {
	return sendProductPayload("POST", payload)
}

func putProduct(payload *data.Product) *httptest.ResponseRecorder {
	return sendProductPayload("PUT", payload)
}

func TestPostProductSuccess(t *testing.T) {
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	deleteItems := insertItems(tplOffer, tplAccess)
	defer deleteItems()
	payload := validProductPayload(tplOffer.ID, tplAccess.ID)
	res := postProduct(&payload)
	if res.Code != http.StatusCreated {
		t.Fatalf("failed to post product: %d", res.Code)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	product := &data.Product{}
	if err := testServer.db.FindByPrimaryKeyTo(product, reply.ID); err != nil {
		t.Fatal("failed to get product: ", err)
	}
	testServer.db.Delete(product)
}

func TestPostProductValidation(t *testing.T) {
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	deleteItems := insertItems(tplOffer, tplAccess)
	defer deleteItems()
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
		res := postProduct(&payload)
		if res.Code != http.StatusBadRequest {
			t.Error("failed validation: ", res.Code)
		}
	}
}

type productTestData struct {
	TplOffer  *data.Template
	TplAccess *data.Template
	Product   *data.Product
}

func createProductTestData() (*productTestData, func()) {
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	tplAccess := data.NewTestTemplate(data.TemplateAccess)
	product := &data.Product{
		ID:            util.NewUUID(),
		Name:          "foo",
		OfferTplID:    &tplOffer.ID,
		OfferAccessID: &tplAccess.ID,
		UsageRepType:  data.ProductUsageTotal,
	}
	deleteItems := insertItems(tplOffer, tplAccess, product)
	return &productTestData{tplOffer, tplAccess, product}, deleteItems
}

func TestPutProduct(t *testing.T) {
	testData, deleteItems := createProductTestData()
	defer deleteItems()
	payload := validProductPayload(testData.TplOffer.ID, testData.TplAccess.ID)
	payload.ID = testData.Product.ID
	res := putProduct(&payload)
	if res.Code != http.StatusOK {
		t.Fatalf("failed to put product: %d", res.Code)
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

func getProducts() *httptest.ResponseRecorder {
	return getResources(productsPath, nil, testServer.handleProducts)
}

func testGetProducts(t *testing.T, exp int) {
	res := getProducts()
	testGetResources(t, res, exp)
}

func TestGetProducts(t *testing.T) {
	testServer.db.DeleteFrom(data.ProductTable, "")
	// Get empty list.
	testGetProducts(t, 0)
	// Get all products.
	_, deleteItems := createProductTestData()
	defer deleteItems()
	testGetProducts(t, 1)
}
