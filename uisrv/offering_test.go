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
	"github.com/privatix/dappctrl/proc/worker"
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

func responseOfferingsArray(t *testing.T,
	res *http.Response, exp int) []*data.Offering {
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to get resources: ", res.StatusCode)
	}
	var ret []*data.Offering
	json.NewDecoder(res.Body).Decode(&ret)
	if exp != len(ret) {
		t.Fatalf("expected %d items, got: %d (%s)", exp, len(ret),
			util.Caller())
	}
	return ret
}

func testSortOfferings(t *testing.T, id, product,
	status string, pattern []string) {
	res := getResources(t, offeringsPath,
		map[string]string{
			"id":          id,
			"product":     product,
			"offerStatus": status})
	offerings := responseOfferingsArray(t, res, len(pattern))

	for k, v := range offerings {
		if v.ID != pattern[k] {
			t.Fatal("offerings sorted incorrectly")
		}
	}
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
	off1.BlockNumberUpdated = 1

	off2 := data.NewTestOffering(testAgent.EthAddr,
		testProd.ID, testTpl.ID)
	off2.OfferStatus = data.OfferEmpty
	off2.BlockNumberUpdated = 2

	off3 := data.NewTestOffering(createNotUsedAcc(t).EthAddr,
		testProd.ID, testTpl.ID)
	off3.OfferStatus = data.OfferRegister

	off4 := data.NewTestOffering(genEthAddr(t),
		testProd.ID, testTpl.ID)
	off4.OfferStatus = data.OfferRegister

	insertItems(t, off1, off2, off3, off4)

	sortPattern := []string{off2.ID, off1.ID}

	// Get all offerings.
	testGetOfferings(t, "", "", "", 2)
	testSortOfferings(t, "", "", "", sortPattern)

	// Get offerings by id.
	testGetOfferings(t, off1.ID, "", "", 1)

	// Get offerings by product.
	testGetOfferings(t, "", testProd.ID, "", 2)

	// Get offerings by status.
	testGetOfferings(t, "", "", data.OfferEmpty, 1)
}

func testGetClientOfferingsOrdered(
	t *testing.T, params map[string]string, exp int) {
	res := getResources(t, clientOfferingsPath, params)
	ret := testGetResources(t, res, exp)

	// Test the order.
	if len(ret) > 0 {
		prevBlock := uint64(ret[0]["blockNumberUpdated"].(float64))
		for _, item := range ret[1:] {
			cur := uint64(item["blockNumberUpdated"].(float64))
			if cur > prevBlock {
				t.Logf("%v > %v", cur, prevBlock)
				t.Fatal("offerings must be ordered by block")
			}
			prevBlock = cur
		}
	}
}

func TestGetClientOffering(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	createOfferingFixtures(t)
	// Get empty list.
	testGetClientOfferingsOrdered(t, nil, 0)

	// Insert test offerings.
	off1 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off1.OfferStatus = data.OfferRegister
	off1.Status = data.MsgChPublished
	off1.IsLocal = false
	off1.Country = "US"
	off1.BlockNumberUpdated = 11

	off2 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off2.OfferStatus = data.OfferRegister
	off2.Status = data.MsgChPublished
	off2.IsLocal = false
	off2.Country = "SU"
	off2.BlockNumberUpdated = 11111

	off3 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off3.OfferStatus = data.OfferEmpty
	off3.IsLocal = false
	off3.Country = "SU"
	off3.BlockNumberUpdated = 111

	off4 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off4.OfferStatus = data.OfferEmpty
	off4.IsLocal = true
	off4.BlockNumberUpdated = 111111

	off5 := data.NewTestOffering(testAgent.EthAddr, testProd.ID,
		testTpl.ID)
	off5.OfferStatus = data.OfferRegister
	off5.IsLocal = false
	off5.Country = "SU"
	off5.BlockNumberUpdated = 2

	off6 := data.NewTestOffering(genEthAddr(t), testProd.ID, testTpl.ID)
	off6.OfferStatus = data.OfferRegister
	off6.Status = data.MsgChPublished
	off6.IsLocal = false
	off6.Country = "US"
	off6.CurrentSupply = 0
	off6.BlockNumberUpdated = 222

	insertItems(t, off1, off2, off3, off4, off5, off6)

	// All non-local offerings
	testGetClientOfferingsOrdered(t, nil, 2)

	lowPrice := strconv.FormatUint(off1.UnitPrice-10, 10)
	price := strconv.FormatUint(off1.UnitPrice, 10)
	highPrice := strconv.FormatUint(off1.UnitPrice+10, 10)

	// Filter by price range
	testGetClientOfferingsOrdered(t, nil, 2) // inside range
	testGetClientOfferingsOrdered(t, map[string]string{
		"maxUnitPrice": highPrice,
	}, 2) // inside range
	testGetClientOfferingsOrdered(t,
		map[string]string{"maxUnitPrice": lowPrice}, 0) // above range
	testGetClientOfferingsOrdered(t,
		map[string]string{"minUnitPrice": highPrice}, 0) // below range
	testGetClientOfferingsOrdered(t, map[string]string{
		"minUnitPrice": lowPrice,
		"maxUnitPrice": price}, 2) // on edge
	testGetClientOfferingsOrdered(t, map[string]string{
		"minUnitPrice": price,
		"maxUnitPrice": highPrice}, 2) // on edge
	testGetClientOfferingsOrdered(t, map[string]string{
		"minUnitPrice": price,
		"maxUnitPrice": price}, 2) // on edge

	// Filter by country
	testGetClientOfferingsOrdered(t,
		map[string]string{"country": "US"}, 1)
	testGetClientOfferingsOrdered(t,
		map[string]string{"country": "SU"}, 1)
	testGetClientOfferingsOrdered(t,
		map[string]string{"country": "US,SU"}, 2)

	// Filter by agent
	testGetClientOfferingsOrdered(t,
		map[string]string{"agent": genEthAddr(t)}, 0)
	testGetClientOfferingsOrdered(t,
		map[string]string{"agent": off1.Agent}, 1)

	// Get offering by id
	testGetClientOfferingsOrdered(t,
		map[string]string{"id": util.NewUUID()}, 0)
	testGetClientOfferingsOrdered(t,
		map[string]string{"id": off6.ID}, 1)
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

func sendOfferingAction(t *testing.T, id, action string,
	gasPrice uint64) *http.Response {
	path := offeringsPath + id + "/status"
	return sendPayload(t, http.MethodPut, path,
		&OfferingPutPayload{Action: action, GasPrice: gasPrice})
}

func sendClientOfferingAction(t *testing.T, id, action,
	account string, gasPrice uint64, deposit uint64) *http.Response {
	path := clientOfferingsPath + id + "/status"
	return sendPayload(t, http.MethodPut, path,
		&ClientOfferingPutPayload{Action: action, Account: account,
			GasPrice: gasPrice, Deposit: deposit})
}

func TestPutOfferingStatus(t *testing.T) {
	fxt := data.NewTestFixture(t, testServer.db)
	defer fxt.Close()
	defer setTestUserCredentials(t)()

	res := sendOfferingAction(t, fxt.Offering.ID, "wrong-action",
		uint64(1))
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("wanted: %d, got: %v",
			http.StatusBadRequest, res.Status)
	}

	testPutOfferingStatusCreatesJob := func(
		action, jobType string, sentGasPrice, expectedGasPrice uint64) {
		res := sendOfferingAction(t, fxt.Offering.ID, action,
			sentGasPrice)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("got: %v (%s)", res.Status, util.Caller())
		}
		job := &data.Job{}
		data.FindInTestDB(t, testServer.db, job, "related_id",
			fxt.Offering.ID)
		defer data.DeleteFromTestDB(t, testServer.db, job)

		if job.Type != jobType {
			t.Fatalf("unexpected job created, wanted: %s, got: %s (%s)",
				jobType, job.Type, util.Caller())
		}

		expectedData, _ := json.Marshal(&data.JobPublishData{
			GasPrice: expectedGasPrice,
		})
		if !bytes.Equal(job.Data, expectedData) {
			t.Fatalf("job does not contain expected data (%s)", util.Caller())
		}
	}

	testPutOfferingStatusCreatesJob(PublishOffering,
		data.JobAgentPreOfferingMsgBCPublish, 1, 1)

	testPutOfferingStatusCreatesJob(PopupOffering,
		data.JobAgentPreOfferingPopUp, 1, 1)

	testPutOfferingStatusCreatesJob(DeactivateOffering,
		data.JobAgentPreOfferingDelete, 1, 1)

	testDefaultGasPrice := uint64(20000000)

	gasPriceSettings := &data.Setting{
		Key:   data.SettingDefaultGasPrice,
		Value: fmt.Sprint(testDefaultGasPrice),
	}
	data.InsertToTestDB(t, testServer.db, gasPriceSettings)
	defer data.DeleteFromTestDB(t, testServer.db, gasPriceSettings)

	testPutOfferingStatusCreatesJob(PopupOffering,
		data.JobAgentPreOfferingPopUp, 0, testDefaultGasPrice)
}

func TestPutClientOfferingStatus(t *testing.T) {
	defer cleanDB(t)

	setTestUserCredentials(t)
	createOfferingFixtures(t)

	testGasPrice := uint64(1)

	offer := data.NewTestOffering(genEthAddr(t),
		testProd.ID, testTpl.ID)
	offer.OfferStatus = data.OfferRegister
	offer.Status = data.MsgChPublished
	offer.IsLocal = false
	offer.Country = "US"

	insertItems(t, offer)

	minDeposit := data.MinDeposit(offer)

	res := sendClientOfferingAction(t, offer.ID, "wrong-action",
		testAgent.ID, testGasPrice, minDeposit)

	checkStatusCode(t, res, http.StatusBadRequest,
		"failed to put offering status: %d")

	res = sendClientOfferingAction(t, offer.ID, AcceptOffering,
		testAgent.ID, testGasPrice, minDeposit)

	checkStatusCode(t, res, http.StatusOK,
		"failed to put offering status: %d")

	res = sendClientOfferingAction(t, offer.ID, AcceptOffering,
		testAgent.ID, testGasPrice, minDeposit-1)

	checkStatusCode(t, res, http.StatusBadRequest,
		"failed to validate custom deposit: %d")

	// Custom deposit is not specified.
	res = sendClientOfferingAction(t, offer.ID, AcceptOffering,
		testAgent.ID, testGasPrice, 0)

	checkStatusCode(t, res, http.StatusOK,
		"custom deposit must not be required")

	expectedData, err := json.Marshal(&worker.ClientPreChannelCreateData{
		GasPrice: testGasPrice, Offering: offer.ID,
		Account: testAgent.ID, Deposit: minDeposit})
	if err != nil {
		t.Fatal(err)
	}

	jobs, err := testServer.db.SelectAllFrom(data.JobTable, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, j := range jobs {
		if job, ok := j.(*data.Job); ok {
			if !bytes.Equal(job.Data, expectedData) {
				t.Fatal("job does not contain expected data")
			}
		}
	}
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
