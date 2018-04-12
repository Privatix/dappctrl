// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

var (
	testTpl   *data.Template
	testProd  *data.Product
	testAgent *data.Account
)

func createOfferingFixtures() func() {
	testTpl = data.NewTestTemplate(data.TemplateAccess)
	testProd = data.NewTestProduct()
	testAgent = data.NewTestAccount()
	return insertItems(testTpl, testProd, testAgent)
}

func validOfferingPayload() data.Offering {
	return data.Offering{
		AdditionalParams:   []byte("{}"),
		Agent:              testAgent.ID,
		BillingInterval:    100,
		BillingType:        data.BillingPrepaid,
		Country:            "KG",
		Description:        nil,
		FreeUnits:          0,
		MaxBillingUnitLag:  100,
		MaxInactiveTimeSec: nil,
		MaxSuspendTime:     1000,
		MaxUnit:            nil,
		MinUnits:           uint64(50),
		Product:            testProd.ID,
		ServiceName:        "my-service",
		SetupPrice:         32,
		Supply:             uint(1),
		Template:           testTpl.ID,
		UnitName:           "Time",
		UnitPrice:          76,
		UnitType:           data.UnitSeconds,
	}
}

func testOfferingReply(t *testing.T, offer *data.Offering) {
	// test hash
	expectedHash := data.FromBytes(data.OfferingHash(offer))
	if offer.Hash != expectedHash {
		t.Errorf("expected hash %s, got: %s", expectedHash, offer.Hash)
	}
	// test signature.
	pub, _ := data.ToBytes(testAgent.PublicKey)
	sig, _ := data.ToBytes(offer.Signature)
	hash := data.OfferingHash(offer)
	if !crypto.VerifySignature(pub, hash, sig[:len(sig)-1]) {
		t.Error("wrong signature")
	}
}

func sendOffering(v *data.Offering, m string) *httptest.ResponseRecorder {
	return sendPayload(m, offeringsPath, v, testServer.handleOfferings)
}

func postOffering(v *data.Offering) *httptest.ResponseRecorder {
	return sendOffering(v, "POST")
}

func putOffering(v *data.Offering) *httptest.ResponseRecorder {
	return sendOffering(v, "PUT")
}

func TestPostOfferingSuccess(t *testing.T) {
	deleteFixtures := createOfferingFixtures()
	defer deleteFixtures()

	// Successful offering creation.
	payload := validOfferingPayload()
	res := postOffering(&payload)
	if res.Code != http.StatusCreated {
		t.Errorf("failed to create, response: %d", res.Code)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	offering := &data.Offering{}
	testServer.db.FindByPrimaryKeyTo(offering, reply.ID)
	testOfferingReply(t, offering)
	testServer.db.Delete(offering)
}

func TestPostOfferingValidation(t *testing.T) {
	// Prepare test data.
	deleteFixtures := createOfferingFixtures()
	defer deleteFixtures()
	validPld := validOfferingPayload()

	invalidUnitType := validPld
	invalidUnitType.UnitType = "Invalid"

	invalidBillingType := validPld
	invalidBillingType.BillingType = "Invalid"

	noAdditionalParams := validPld
	noAdditionalParams.AdditionalParams = nil

	noAgent := validPld
	noAgent.Agent = ""

	noBillingInterval := validPld
	noBillingInterval.BillingInterval = 0

	noBillingType := validPld
	noBillingType.BillingType = ""

	noCountry := validPld
	noCountry.Country = ""

	noMinUnits := validPld
	noMinUnits.MinUnits = 0

	noProduct := validPld
	noProduct.Product = ""

	noServiceName := validPld
	noServiceName.ServiceName = ""

	noSupply := validPld
	noSupply.Supply = 0

	noTemplate := validPld
	noTemplate.Template = ""

	noUnitName := validPld
	noUnitName.UnitName = ""

	noUnitType := validPld
	noUnitType.UnitType = ""

	for _, payload := range []data.Offering{
		invalidUnitType,
		invalidBillingType,

		// Test required fields.
		noAdditionalParams,
		noAgent,
		noBillingInterval,
		noBillingType,
		noCountry,
		noMinUnits,
		noProduct,
		noServiceName,
		noSupply,
		noTemplate,
		noUnitName,
		noUnitType,
	} {
		res := postOffering(&payload)
		if res.Code != http.StatusBadRequest {
			t.Errorf("failed with response: %d", res.Code)
		}
	}
}

func TestPutOfferingSuccess(t *testing.T) {
	deleteFixtures := createOfferingFixtures()
	defer deleteFixtures()
	testOffering := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	deleteOffering := insertItems(testOffering)
	defer deleteOffering()

	// Successful offering creation.
	payload := validOfferingPayload()
	payload.ID = testOffering.ID
	res := putOffering(&payload)
	if res.Code != http.StatusOK {
		t.Fatalf("failed to put, response: %d", res.Code)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	offering := &data.Offering{}
	testServer.db.FindByPrimaryKeyTo(offering, reply.ID)
	testOfferingReply(t, offering)
	testServer.db.Delete(offering)
}

func getOfferings(id string) *httptest.ResponseRecorder {
	return getResources(offeringsPath,
		map[string]string{"id": id},
		testServer.handleOfferings)
}

func testGetOfferings(t *testing.T, id string, exp int) {
	res := getOfferings(id)
	testGetResources(t, res, exp)
}

func TestGetOffering(t *testing.T) {
	deleteFixtures := createOfferingFixtures()
	defer deleteFixtures()
	// Get empty list.
	testGetOfferings(t, "", 0)

	// Get all offerings.
	testOfferings := []*data.Offering{
		data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID),
		data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID),
	}
	deleteOfferings := insertItems(testOfferings[0], testOfferings[1])
	defer deleteOfferings()
	testGetOfferings(t, "", 2)

	// Get offering by id.
	testGetOfferings(t, testOfferings[0].ID, 1)
}

func sendToOfferingStatus(id, action, method string) *httptest.ResponseRecorder {
	path := fmt.Sprintf("%s%s/status", offeringsPath, id)
	var r *http.Request
	if action != "" {
		r.Form = make(url.Values)
		r.Form.Add("action", action)
	}
	r = httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	testServer.handleOfferings(w, r)
	return w
}

func putOfferingStatus(id, action string) *httptest.ResponseRecorder {
	return sendToOfferingStatus(id, action, "PUT")
}

func getOfferingStatus(id string) *httptest.ResponseRecorder {
	return sendToOfferingStatus(id, "", "GET")
}

func TestPutOfferingStatus(t *testing.T) {
	// TODO once job queue implemented.
}

func TestGetOfferingStatus(t *testing.T) {
	deleteFixtures := createOfferingFixtures()
	defer deleteFixtures()
	testOffering := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	deleteOffering := insertItems(testOffering)
	defer deleteOffering()
	// Get offering status with a match.
	res := getOfferingStatus(testOffering.ID)
	if res.Code != http.StatusOK {
		t.Fatalf("failed to get status: %d", res.Code)
	}
	reply := &statusReply{}
	json.NewDecoder(res.Body).Decode(reply)
	if testOffering.Status != reply.Status {
		t.Fatalf("expected %s, got: %s", testOffering.Status, reply.Status)
	}
	// Get offering status without a match.
	res = getOfferingStatus(util.NewUUID())
	if res.Code != http.StatusNotFound {
		t.Fatalf("expected not found, got: %d", res.Code)
	}
}
