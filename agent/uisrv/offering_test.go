// +build !noagentuisrvtest

package uisrv

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func createOfferingFixtures() {
	testTpl = data.NewTestTemplate(data.TemplateAccess)
	testProd = data.NewTestProduct()
	testAgent = data.NewTestAccount()
	insertItems(testTpl, testProd, testAgent)
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

func sendOffering(t *testing.T, v *data.Offering, m string) *http.Response {
	return sendPayload(t, m, offeringsPath, v)
}

func postOffering(t *testing.T, v *data.Offering) *http.Response {
	return sendOffering(t, v, "POST")
}

func putOffering(t *testing.T, v *data.Offering) *http.Response {
	return sendOffering(t, v, "PUT")
}

func TestPostOfferingSuccess(t *testing.T) {
	defer cleanDB()

	createOfferingFixtures()

	// Successful offering creation.
	payload := validOfferingPayload()
	res := postOffering(t, &payload)
	if res.StatusCode != http.StatusCreated {
		t.Errorf("failed to create, response: %d", res.StatusCode)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	offering := &data.Offering{}
	testServer.db.FindByPrimaryKeyTo(offering, reply.ID)
	testOfferingReply(t, offering)
}

func TestPostOfferingValidation(t *testing.T) {
	defer cleanDB()
	// Prepare test data.
	createOfferingFixtures()
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
		res := postOffering(t, &payload)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("failed with response: %d", res.StatusCode)
		}
	}
}

func TestPutOfferingSuccess(t *testing.T) {
	defer cleanDB()

	createOfferingFixtures()
	testOffering := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	insertItems(testOffering)

	// Successful offering creation.
	payload := validOfferingPayload()
	payload.ID = testOffering.ID
	res := putOffering(t, &payload)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("failed to put, response: %d", res.StatusCode)
	}
	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	offering := &data.Offering{}
	testServer.db.FindByPrimaryKeyTo(offering, reply.ID)
	testOfferingReply(t, offering)
}

func getOfferings(t *testing.T, id string) *http.Response {
	return getResources(t, offeringsPath,
		map[string]string{"id": id})
}

func testGetOfferings(t *testing.T, id string, exp int) {
	res := getOfferings(t, id)
	testGetResources(t, res, exp)
}

func TestGetOffering(t *testing.T) {
	defer cleanDB()

	createOfferingFixtures()
	// Get empty list.
	testGetOfferings(t, "", 0)

	// Get all offerings.
	testOfferings := []*data.Offering{
		data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID),
		data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID),
	}
	insertItems(testOfferings[0], testOfferings[1])
	testGetOfferings(t, "", 2)

	// Get offering by id.
	testGetOfferings(t, testOfferings[0].ID, 1)
}

func sendToOfferingStatus(t *testing.T, id, action, method string) *http.Response {
	url := fmt.Sprintf("http://%s%s%s/status",
		testServer.conf.Addr, offeringsPath, id)
	if action != "" {
		url += "?action=" + action
	}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal("failed to create a request: ", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("failed to perform a request: ", err)
	}
	return res
}

func putOfferingStatus(t *testing.T, id, action string) *http.Response {
	return sendToOfferingStatus(t, id, action, "PUT")
}

func getOfferingStatus(t *testing.T, id string) *http.Response {
	return sendToOfferingStatus(t, id, "", "GET")
}

func TestPutOfferingStatus(t *testing.T) {
	// TODO once job queue implemented.
}

func TestGetOfferingStatus(t *testing.T) {
	defer cleanDB()

	createOfferingFixtures()
	offer := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	insertItems(offer)
	// Get offering status with a match.
	res := getOfferingStatus(t, offer.ID)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("failed to get status: %d", res.StatusCode)
	}
	reply := &statusReply{}
	json.NewDecoder(res.Body).Decode(reply)
	if offer.Status != reply.Status {
		t.Fatalf("expected %s, got: %s", offer.Status, reply.Status)
	}
	// Get offering status without a match.
	res = getOfferingStatus(t, util.NewUUID())
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected not found, got: %d", res.StatusCode)
	}
}
