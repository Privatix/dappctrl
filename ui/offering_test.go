package ui_test

import (
	"encoding/json"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type testOfferingData struct {
	agent              string
	offerStatus        string
	status             string
	country            string
	isLocal            bool
	blockNumberUpdated uint64
	currentSupply      uint16
}

type testGetAgentOfferingsArgs struct {
	expErr      error
	expNumber   int
	product     string
	offerStatus string
}

type testGetClientOfferingsArgs struct {
	expErr       error
	expNumber    int
	agent        string
	minUnitPrice uint64
	maxUnitPrice uint64
	country      []string
}

type testField struct {
	field string
	value interface{}
}

func TestAcceptOffering(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "AcceptOffering")
	defer fxt.close()

	var j *data.Job
	handler.SetMockQueue(job.QueueMock(func(method int, j2 *data.Job,
		relatedIDs []string, subID string, subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	}))

	minDeposit := data.MinDeposit(fxt.Offering)

	_, err := handler.AcceptOffering("wrong-password", fxt.UserAcc.ID,
		fxt.Offering.ID, minDeposit, 12345)
	assertMatchErr(ui.ErrAccessDenied, err)

	_, err = handler.AcceptOffering(data.TestPassword, util.NewUUID(),
		fxt.Offering.ID, minDeposit, 12345)
	assertMatchErr(ui.ErrAccountNotFound, err)

	_, err = handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		util.NewUUID(), minDeposit, 12345)
	assertMatchErr(ui.ErrOfferingNotFound, err)

	_, err = handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		fxt.Offering.ID, minDeposit-1, 12345)
	assertMatchErr(ui.ErrDepositTooSmall, err)

	res, err := handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		fxt.Offering.ID, minDeposit, 12345)
	assertMatchErr(nil, err)

	if res == nil || j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != *res ||
		j.Type != data.JobClientPreChannelCreate {
		t.Fatalf("wrong result data")
	}
}

func checkGetClientOfferings(t *testing.T, expErr error, expNumber int,
	agent string, minUnitPrice, maxUnitPrice uint64, country []string,
	checkFunc func(error, error)) {
	res, err := handler.GetClientOfferings(
		data.TestPassword, agent, minUnitPrice, maxUnitPrice, country)
	checkFunc(expErr, err)
	testGetOfferingsResult(t, res, expNumber)
}

func checkGetAgentOfferings(t *testing.T, expErr error, expNumber int, product,
	status string, checkFunc func(error, error)) {
	res, err := handler.GetAgentOfferings(
		data.TestPassword, product, status)
	checkFunc(expErr, err)
	testGetOfferingsResult(t, res, expNumber)
}

func testGetOfferingsResult(
	t *testing.T, res []data.Offering, expNumber int) {
	if res == nil && len(res) == 0 {
		if expNumber > 0 {
			t.Fatalf("number of offerings"+
				" expected: %d, got: 0", expNumber)
		}
		return
	}

	if len(res) != expNumber {
		t.Fatalf("number of offerings expected: %d, got: %d",
			expNumber, len(res))
	}

	if len(res) > 0 {
		prevBlock := res[0].BlockNumberUpdated

		for _, item := range res {
			current := item.BlockNumberUpdated
			if current > prevBlock {
				t.Fatalf("offerings must be ordered by block")
			}
		}
	}
}

func createTestOffering(fxt *fixture, agent, offerStatus, status,
	country string, isLocal bool, blockNumberUpdated uint64,
	currentSupply uint16) *data.Offering {
	offering := data.NewTestOffering(
		agent, fxt.Product.ID, fxt.TemplateOffer.ID)
	if offerStatus != "" {
		offering.OfferStatus = offerStatus
	}

	if status != "" {
		offering.Status = status
	}

	if country != "" {
		offering.Country = country
	}

	offering.IsLocal = isLocal

	if blockNumberUpdated != 0 {
		offering.BlockNumberUpdated = blockNumberUpdated
	}

	offering.CurrentSupply = currentSupply

	return offering
}

func testGetClientOfferings(t *testing.T,
	fxt *fixture, assertMatchErr func(error, error), agent string) {
	_, err := handler.GetClientOfferings(
		"wrong-password", "", 0, 0, nil)
	assertMatchErr(ui.ErrAccessDenied, err)

	lowPrice := fxt.Offering.UnitPrice - 10
	price := fxt.Offering.UnitPrice
	highPrice := fxt.Offering.UnitPrice + 10

	testArgs := []testGetClientOfferingsArgs{
		{nil, 3, "", 0, 0, nil},
		{nil, 0, "", 0, lowPrice, nil},
		{nil, 0, "", highPrice, 0, nil},
		{nil, 3, "", lowPrice, price, nil},
		{nil, 3, "", price, highPrice, nil},
		{nil, 3, "", price, price, nil},
		{ui.ErrBadUnitPriceRange, 0, "", highPrice, lowPrice, nil},
		{nil, 1, "", 0, 0, []string{"US"}},
		{nil, 1, "", 0, 0, []string{"SU"}},
		{nil, 2, "", 0, 0, []string{"SU", "US"}},
		{nil, 0, data.NewTestAccount(data.TestPassword).ID, 0, 0, nil},
		{nil, 1, agent, 0, 0, nil},
	}

	for _, v := range testArgs {
		checkGetClientOfferings(t,
			v.expErr, v.expNumber, v.agent, v.minUnitPrice,
			v.maxUnitPrice, v.country, assertMatchErr)
	}
}

func TestGetClientOfferings(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetClientOfferings")
	defer fxt.close()

	other := data.NewTestAccount(data.TestPassword)
	agent := data.NewTestAccount(data.TestPassword)

	var offerings []reform.Record

	testData := []testOfferingData{
		{agent.EthAddr, data.OfferRegister, data.MsgChPublished,
			"US", false, 11, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferRegister, data.MsgChPublished,
			"SU", false, 11111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty, "",
			"SU", false, 111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty, "",
			"", true, 111111, fxt.Offering.CurrentSupply},
		{agent.EthAddr, data.OfferRegister, "",
			"SU", false, 2, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferRegister, data.MsgChPublished,
			"US", false, 222, 0},
	}

	for _, v := range testData {
		offering := createTestOffering(fxt,
			v.agent, v.offerStatus, v.status, v.country,
			v.isLocal, v.blockNumberUpdated, v.currentSupply)
		offerings = append(offerings, offering)
	}

	for _, offering := range offerings {
		data.InsertToTestDB(t, db, offering)

	}

	defer data.DeleteFromTestDB(t, db, offerings...)

	testGetClientOfferings(t, fxt, assertMatchErr, agent.EthAddr)
}

func testGetAgentOfferings(t *testing.T,
	fxt *fixture, assertMatchErr func(error, error)) {
	_, err := handler.GetAgentOfferings("wrong-password", "", "")
	assertMatchErr(ui.ErrAccessDenied, err)

	testArgs := []testGetAgentOfferingsArgs{
		{nil, 2, "", ""},
		{nil, 2, fxt.Product.ID, ""},
		{nil, 1, "", data.OfferEmpty},
	}

	for _, v := range testArgs {
		checkGetAgentOfferings(t, v.expErr, v.expNumber,
			v.product, v.offerStatus, assertMatchErr)
	}
}

func TestGetAgentOfferings(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetAgentOfferings")
	defer fxt.close()

	createNotUsedAcc := func(t *testing.T) *data.Account {
		acc := data.NewTestAccount(data.TestPassword)
		acc.InUse = false
		data.InsertToTestDB(t, db, acc)
		return acc
	}

	var offerings []reform.Record

	testData := []testOfferingData{
		{fxt.Account.EthAddr, data.OfferRegister, "", "",
			false, 1, fxt.Offering.CurrentSupply},
		{fxt.Account.EthAddr, data.OfferEmpty, "", "",
			false, 2, fxt.Offering.CurrentSupply},
		{createNotUsedAcc(t).EthAddr, data.OfferRegister, "", "",
			false, 2, fxt.Offering.CurrentSupply},
		{createNotUsedAcc(t).EthAddr, data.OfferRegister, "", "",
			false, 2, fxt.Offering.CurrentSupply},
	}

	for _, v := range testData {
		offering := createTestOffering(fxt,
			v.agent, v.offerStatus, v.status, v.country,
			v.isLocal, v.blockNumberUpdated, v.currentSupply)
		offerings = append(offerings, offering)
	}

	for _, offering := range offerings {
		data.InsertToTestDB(t, db, offering)

	}

	defer data.DeleteFromTestDB(t, db, offerings...)

	testGetAgentOfferings(t, fxt, assertMatchErr)
}

func genOffering(t *testing.T,
	fxt *fixture, field string, value interface{}) *data.Offering {
	tempOffering := data.NewTestOffering(
		fxt.Account.EthAddr, fxt.Product.ID, fxt.TemplateOffer.ID)

	raw, err := json.Marshal(tempOffering)
	if err != nil {
		t.Fatal(err)
	}

	unpackedFields := make(map[string]interface{})

	err = json.Unmarshal(raw, &unpackedFields)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := unpackedFields[field]; !ok {
		t.Fatal("field not found")
	}

	unpackedFields[field] = value

	raw2, err := json.Marshal(unpackedFields)
	if err != nil {
		t.Fatal(err)
	}

	var offering *data.Offering
	err = json.Unmarshal(raw2, &offering)
	if err != nil {
		t.Fatal(err)
	}

	return offering
}

func invalidOfferingsArray(
	t *testing.T, fxt *fixture) (offerings []*data.Offering) {

	testFields := []testField{
		{"unitType", "Invalid"},
		{"billingType", "Invalid"},
		{"additionalParams", nil},
		{"agent", ""},
		{"agent", ""},
		{"billingInterval", 0},
		{"billingType", ""},
		{"country", ""},
		{"minUnits", 0},
		{"product", ""},
		{"serviceName", ""},
		{"supply", 0},
		{"template", ""},
		{"unitName", ""},
		{"unitType", ""},
	}

	for _, v := range testFields {
		offering := genOffering(t, fxt, v.field, v.value)
		offerings = append(offerings, offering)
	}

	return offerings
}

func TestCreateOffering(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "CreateOffering")
	defer fxt.close()

	invalidOfferings := invalidOfferingsArray(t, fxt)

	for _, v := range invalidOfferings {
		_, err := handler.CreateOffering(data.TestPassword, v)
		if err == nil {
			t.Fatal("offering should not be saved")
		}
	}

	offering := data.NewTestOffering(
		fxt.Account.ID, fxt.Product.ID, fxt.TemplateOffer.ID)

	_, err := handler.CreateOffering("wrong-password", offering)
	assertMatchErr(ui.ErrAccessDenied, err)

	res, err := handler.CreateOffering(data.TestPassword, offering)
	assertMatchErr(nil, err)

	offering2 := &data.Offering{}
	err = db.FindByPrimaryKeyTo(offering2, res)
	assertMatchErr(nil, err)

	data.DeleteFromTestDB(t, db, offering2)
}

func TestUpdateOffering(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "UpdateOffering")
	defer fxt.close()

	err := handler.UpdateOffering("wrong-password", fxt.Offering)
	assertMatchErr(ui.ErrAccessDenied, err)

	newOffering := data.NewTestOffering(
		fxt.Account.ID, fxt.Product.ID, fxt.TemplateOffer.ID)

	err = handler.UpdateOffering(data.TestPassword, newOffering)
	assertMatchErr(ui.ErrOfferingNotFound, err)

	fxt.Offering.Status = data.MsgChPublished

	err = handler.UpdateOffering(data.TestPassword, fxt.Offering)
	assertMatchErr(nil, err)

	savedOffering := &data.Offering{}
	err = db.FindByPrimaryKeyTo(savedOffering, fxt.Offering.ID)
	if err != nil {
		t.Fatal(err)
	}

	if savedOffering.Status != data.MsgChPublished {
		t.Fatal("offering not updated")
	}
}

func TestChangeOfferingStatus(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "ChangeOfferingStatus")
	defer fxt.close()

	var j *data.Job
	handler.SetMockQueue(job.QueueMock(func(method int, j2 *data.Job,
		relatedIDs []string, subID string, subFunc job.SubFunc) error {
		switch method {
		case job.MockAdd:
			j = j2
		default:
			t.Fatal("unexpected queue call")
		}
		return nil
	}))

	for action, jobType := range ui.OfferingChangeActions {
		err := handler.ChangeOfferingStatus(
			data.TestPassword, fxt.Offering.ID, action, 100)
		assertMatchErr(nil, err)

		if j == nil || j.Type != jobType ||
			j.RelatedID != fxt.Offering.ID {
			t.Fatal("expected job not created")
		}
	}
}
