// +build !noagentuisrvtest

package uisrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

var (
	testTpl   *data.Template
	testProd  *data.Product
	testAgent *data.Account
)

func createOfferingFixtures(t *testing.T) {
	testTpl = data.NewTestTemplate(data.TemplateAccess)
	testProd = data.NewTestProduct()
	testAgent = data.NewTestAccount(testPassword)
	insertItems(t, testTpl, testProd, testAgent)
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
		Supply:             1,
		Template:           testTpl.ID,
		UnitName:           "Time",
		UnitPrice:          76,
		UnitType:           data.UnitSeconds,
	}
}

func sendOffering(t *testing.T, v *data.Offering, m string) *http.Response {
	return sendPayload(t, m, offeringsPath, v)
}

func postOffering(t *testing.T, v *data.Offering) *http.Response {
	return sendOffering(t, v, http.MethodPost)
}

func putOffering(t *testing.T, v *data.Offering) *http.Response {
	return sendOffering(t, v, "PUT")
}

func TestPostOfferingSuccess(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createOfferingFixtures(t)

	// Successful offering creation.
	payload := validOfferingPayload()
	res := postOffering(t, &payload)
	if res.StatusCode != http.StatusCreated {
		t.Errorf("failed to create, response: %d", res.StatusCode)
	}
}

func TestPostOfferingValidation(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Prepare test data.
	createOfferingFixtures(t)
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
	defer cleanDB(t)
	setTestUserCredentials(t)

	createOfferingFixtures(t)
	testOffering := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	insertItems(t, testOffering)

	// Successful offering creation.
	payload := validOfferingPayload()
	payload.ID = testOffering.ID
	res := putOffering(t, &payload)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("failed to put, response: %d", res.StatusCode)
	}
}

func testGetOfferings(t *testing.T, id, product, status string, exp int) {
	res := getResources(t, offeringsPath,
		map[string]string{
			"id":          id,
			"product":     product,
			"offerStatus": status})
	testGetResources(t, res, exp)
}

func TestGetOffering(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createNotUsedAcc := func(t *testing.T) *data.Account {
		acc := data.NewTestAccount(testPassword)
		acc.InUse = false
		insertItems(t, acc)
		return acc
	}

	createOfferingFixtures(t)
	// Get empty list.
	testGetOfferings(t, "", "", "", 0)

	// Insert test offerings.
	off1 := data.NewTestOffering(testAgent.EthAddr,
		testProd.ID, testTpl.ID)
	off1.OfferStatus = data.OfferRegister

	off2 := data.NewTestOffering(testAgent.EthAddr,
		testProd.ID, testTpl.ID)
	off2.OfferStatus = data.OfferEmpty

	off3 := data.NewTestOffering(createNotUsedAcc(t).EthAddr,
		testProd.ID, testTpl.ID)
	off3.OfferStatus = data.OfferRegister

	off4 := data.NewTestOffering(genEthAddr(t),
		testProd.ID, testTpl.ID)
	off4.OfferStatus = data.OfferRegister

	insertItems(t, off1, off2, off3, off4)

	// Get all offerings.
	testGetOfferings(t, "", "", "", 2)

	// Get offerings by id.
	testGetOfferings(t, off1.ID, "", "", 1)

	// Get offerings by product.
	testGetOfferings(t, "", testProd.ID, "", 2)

	// Get offerings by status.
	testGetOfferings(t, "", "", data.OfferEmpty, 1)
}

func testGetClientOfferings(t *testing.T, minp, maxp, country string, exp int) {
	res := getResources(t, clientOfferingsPath,
		map[string]string{
			"minUnitPrice": minp,
			"maxUnitPrice": maxp,
			"country":      country})
	testGetResources(t, res, exp)
}

func TestGetClientOffering(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createOfferingFixtures(t)
	// Get empty list.
	testGetClientOfferings(t, "", "", "", 0)

	// Insert test offerings.
	off1 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off1.OfferStatus = data.OfferRegister
	off1.Status = data.MsgChPublished
	off1.IsLocal = false
	off1.Country = "US"

	off2 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off2.OfferStatus = data.OfferRegister
	off2.Status = data.MsgChPublished
	off2.IsLocal = false
	off2.Country = "SU"

	off3 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off3.OfferStatus = data.OfferEmpty
	off3.IsLocal = false
	off3.Country = "SU"

	off4 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off4.OfferStatus = data.OfferEmpty
	off4.IsLocal = true

	off5 := data.NewTestOffering(testAgent.EthAddr, testProd.ID,
		testTpl.ID)
	off5.OfferStatus = data.OfferRegister
	off5.IsLocal = false
	off5.Country = "SU"

	insertItems(t, off1, off2, off3, off4, off5)

	// all non-local offerings
	testGetClientOfferings(t, "", "", "", 2)

	lowPrice := strconv.FormatUint(off1.UnitPrice-10, 10)
	price := strconv.FormatUint(off1.UnitPrice, 10)
	highPrice := strconv.FormatUint(off1.UnitPrice+10, 10)

	// price range
	testGetClientOfferings(t, "", "", "", 2)           // inside range
	testGetClientOfferings(t, "", highPrice, "", 2)    // inside range
	testGetClientOfferings(t, "", lowPrice, "", 0)     // above range
	testGetClientOfferings(t, highPrice, "", "", 0)    // below range
	testGetClientOfferings(t, lowPrice, price, "", 2)  // on edge
	testGetClientOfferings(t, price, highPrice, "", 2) // on edge
	testGetClientOfferings(t, price, price, "", 2)     // on edge

	// country filter
	testGetClientOfferings(t, "", "", "US", 1)
	testGetClientOfferings(t, "", "", "SU", 1)
	testGetClientOfferings(t, "", "", "US,SU", 2)
}

func getOfferingStatus(t *testing.T, id string) *http.Response {
	url := fmt.Sprintf("http://:%s@%s%s%s/status", testPassword,
		testServer.conf.Addr, offeringsPath, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

func sendOfferingAction(t *testing.T, id, action string, gasPrice uint64) *http.Response {
	path := fmt.Sprint(offeringsPath, id, "/status")
	payload := &OfferingPutPayload{Action: action, GasPrice: gasPrice}
	return sendPayload(t, http.MethodPut, path, payload)
}

func TestPutOfferingStatus(t *testing.T) {
	fixture := data.NewTestFixture(t, testServer.db)
	defer fixture.Close()
	defer setTestUserCredentials(t)()

	testGasPrice := uint64(1)

	res := sendOfferingAction(t, fixture.Offering.ID, "wrong-action", testGasPrice)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("wanted: %d, got: %v", http.StatusBadRequest, res.Status)
	}

	res = sendOfferingAction(t, fixture.Offering.ID, PublishOffering, testGasPrice)
	if res.StatusCode != http.StatusOK {
		t.Fatal("got: ", res.Status)
	}
	jobPublish := &data.Job{}
	data.FindInTestDB(t, testServer.db, jobPublish, "related_id",
		fixture.Offering.ID)
	expectedData, _ := json.Marshal(&data.JobPublishData{
		GasPrice: testGasPrice,
	})
	if !bytes.Equal(jobPublish.Data, expectedData) {
		t.Fatal("job does not contain expected data")
	}
	data.DeleteFromTestDB(t, testServer.db, jobPublish)
}

func TestGetOfferingStatus(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createOfferingFixtures(t)
	offer := data.NewTestOffering(testAgent.EthAddr, testProd.ID, testTpl.ID)
	insertItems(t, offer)
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
