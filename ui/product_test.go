package ui_test

import (
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
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

func TestCrateProduct(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "CreateProduct")
	defer fxt.close()

	product := testProduct(fxt.TemplateOffer.ID, fxt.TemplateAccess.ID)

	_, err := handler.CreateProduct("wrong-password", product)
	assertErrEqual(ui.ErrAccessDenied, err)

	res, err := handler.CreateProduct(data.TestPassword, product)

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

	err := handler.UpdateProduct("wrong-password", product)
	assertErrEqual(ui.ErrAccessDenied, err)

	unknownProduct := data.Product{ID: util.NewUUID()}
	err = handler.UpdateProduct(data.TestPassword, unknownProduct)
	assertErrEqual(ui.ErrProductNotFound, err)

	assertErrEqual(nil, handler.UpdateProduct(data.TestPassword, product))
	fxt.DB.Reload(&product)
	if product.Name != newName ||
		product.Salt == 0 || product.Password == "" {
		t.Fatal("product was not updated properly")
	}
}

func TestGetProducts(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetProducts")
	defer fxt.close()

	_, err := handler.GetProducts("wrong-password")
	assertErrEqual(ui.ErrAccessDenied, err)

	// pr2 expected to be ignored from reply.
	pr2 := *fxt.Product
	pr2.ID = util.NewUUID()
	pr2.IsServer = false
	fxt.Product.IsServer = true
	data.SaveToTestDB(t, fxt.DB, fxt.Product)
	data.InsertToTestDB(t, fxt.DB, &pr2)
	defer data.DeleteFromTestDB(t, fxt.DB, &pr2)

	result, _ := handler.GetProducts(data.TestPassword)
	if len(result) != 1 {
		t.Fatal("expected 1 product, got ", len(result))
	}
}
