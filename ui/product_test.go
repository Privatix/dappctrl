package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

var (
	badHosts = []string{
		"256.0.0.1",
		"127.256.0.1",
		"127.0.256.1",
		"127.0.0.256",
		"127.0.0.",
		"127.0.0.0.0",
		".127.0.0.1",
		"127.0.0.1.",
		"127.0.0..1",
		"127.0..0.1",
		"127.0..0.1",
		"127..0.0.1",
		"+example.com",
		".example.com",
		"!fff!.dfd.ff",
		"mail@com",
	}

	goodHosts = []string{
		"127.0.0.1",
		"privatix.io",
	}
)

func testProduct(offerTpl, accessTpl string) data.Product {
	return data.Product{
		Name:          "Test product",
		UsageRepType:  data.ProductUsageTotal,
		Salt:          data.TestSalt,
		Password:      data.TestPasswordHash,
		ClientIdent:   data.ClientIdentByChannelID,
		Config:        []byte("{}"),
		OfferTplID:    &offerTpl,
		OfferAccessID: &accessTpl,
	}
}

func checkSEAddress(product *data.Product,
	actionFunc func(password string, product data.Product) error,
	hosts []string, exp error,
	assertErrEqual func(error, error)) {
	for _, host := range hosts {
		product.ServiceEndpointAddress = &host
		err := actionFunc(testToken.v, *product)
		assertErrEqual(exp, err)
	}
}

func TestCrateProduct(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "CreateProduct")
	defer fxt.close()

	product := testProduct(fxt.TemplateOffer.ID, fxt.TemplateAccess.ID)

	_, err := handler.CreateProduct("wrong-token", product)
	assertErrEqual(ui.ErrAccessDenied, err)

	actionFunc := func(password string, product data.Product) error {
		_, err := handler.CreateProduct(password, product)
		return err
	}

	checkSEAddress(&product, actionFunc, badHosts,
		ui.ErrBadServiceEndpointAddress, assertErrEqual)

	product.ServiceEndpointAddress = nil

	res, err := handler.CreateProduct(testToken.v, product)

	prodInDB := &data.Product{}
	err = fxt.DB.FindByPrimaryKeyTo(prodInDB, res)
	if err != nil {
		t.Fatalf("failed to find created product: %v", err)
	}

	data.DeleteFromTestDB(t, fxt.DB, prodInDB)
}

func TestUpdateProduct(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "UpdateProduct")
	defer fxt.close()

	newName := "changed-name"
	product := *fxt.Product
	product.Name = newName
	product.Salt = 0
	product.Password = ""

	err := handler.UpdateProduct("wrong-token", product)
	assertErrEqual(ui.ErrAccessDenied, err)

	checkSEAddress(&product, handler.UpdateProduct, badHosts,
		ui.ErrBadServiceEndpointAddress, assertErrEqual)
	checkSEAddress(&product, handler.UpdateProduct, goodHosts,
		nil, assertErrEqual)

	product.ServiceEndpointAddress = nil

	unknownProduct := data.Product{ID: util.NewUUID()}
	err = handler.UpdateProduct(testToken.v, unknownProduct)
	assertErrEqual(ui.ErrProductNotFound, err)

	assertErrEqual(nil, handler.UpdateProduct(testToken.v, product))
	fxt.DB.Reload(&product)
	if product.Name != newName ||
		product.Salt == 0 || product.Password == "" {
		t.Fatal("product was not updated properly")
	}
}

func TestGetProducts(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetProducts")
	defer fxt.close()

	_, err := handler.GetProducts("wrong-token")
	assertErrEqual(ui.ErrAccessDenied, err)

	// pr2 expected to be ignored from reply.
	pr2 := *fxt.Product
	pr2.ID = util.NewUUID()
	pr2.IsServer = false
	fxt.Product.IsServer = true
	data.SaveToTestDB(t, fxt.DB, fxt.Product)
	data.InsertToTestDB(t, fxt.DB, &pr2)
	defer data.DeleteFromTestDB(t, fxt.DB, &pr2)

	result, _ := handler.GetProducts(testToken.v)
	if len(result) != 1 {
		t.Fatal("expected 1 product, got ", len(result))
	}
}
