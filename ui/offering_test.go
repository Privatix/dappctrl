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
	exp         int
	product     string
	offerStatus string
	offset      uint
	limit       uint
	total       int
}

type testGetClientOfferingsArgs struct {
	exp          int
	agent        string
	minUnitPrice uint64
	maxUnitPrice uint64
	country      []string
	offset       uint
	limit        uint
	total        int
}

type testField struct {
	field string
	value interface{}
}

func TestAcceptOffering(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "AcceptOffering")
	defer fxt.close()

	var j *data.Job
	handler.SetMockQueue(job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
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
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.AcceptOffering(data.TestPassword, util.NewUUID(),
		fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(ui.ErrAccountNotFound, err)

	_, err = handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		util.NewUUID(), minDeposit, 12345)
	assertErrEqual(ui.ErrOfferingNotFound, err)

	_, err = handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		fxt.Offering.ID, minDeposit-1, 12345)
	assertErrEqual(ui.ErrDepositTooSmall, err)

	res, err := handler.AcceptOffering(data.TestPassword, fxt.UserAcc.ID,
		fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(nil, err)

	if res == nil || j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != *res ||
		j.Type != data.JobClientPreChannelCreate {
		t.Fatalf("wrong result data")
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
	fxt *fixture, assertErrEqual func(error, error), agent string) {

	assertResult := func(res *ui.GetClientOfferingsResult,
		err error, exp, total int) {
		assertErrEqual(nil, err)
		if res == nil {
			t.Fatal("result is empty")
		}
		if len(res.Items) != exp {
			t.Fatalf("wanted: %v, got: %v", exp, len(res.Items))
		}
		if res.TotalItems != total {
			t.Fatalf("wanted: %v, got: %v", total, res.TotalItems)
		}
	}

	lowPrice := fxt.Offering.UnitPrice - 10
	price := fxt.Offering.UnitPrice
	highPrice := fxt.Offering.UnitPrice + 10

	_, err := handler.GetClientOfferings(
		"wrong-password", "", 0, 0, nil, 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.GetClientOfferings(
		data.TestPassword, "", highPrice, lowPrice, nil, 0, 0)
	assertErrEqual(ui.ErrBadUnitPriceRange, err)

	other := data.NewTestAccount(data.TestPassword).ID

	testArgs := []testGetClientOfferingsArgs{
		// Test pagination.
		{1, "", 0, 0, nil, 0, 1, 3},
		{2, "", 0, 0, nil, 1, 3, 3},
		{0, "", 0, 0, nil, 3, 3, 3},
		// Test by filters.
		{3, "", 0, 0, nil, 0, 0, 3},
		{0, "", 0, lowPrice, nil, 0, 0, 0},
		{0, "", highPrice, 0, nil, 0, 0, 0},
		{3, "", lowPrice, price, nil, 0, 0, 3},
		{3, "", price, highPrice, nil, 0, 0, 3},
		{3, "", price, price, nil, 0, 0, 3},
		{1, "", 0, 0, []string{"US"}, 0, 0, 1},
		{1, "", 0, 0, []string{"SU"}, 0, 0, 1},
		{2, "", 0, 0, []string{"SU", "US"}, 0, 0, 2},
		{0, other, 0, 0, nil, 0, 0, 0},
		{1, agent, 0, 0, nil, 0, 0, 1},
	}

	for _, v := range testArgs {
		res, err := handler.GetClientOfferings(data.TestPassword,
			v.agent, v.minUnitPrice, v.maxUnitPrice, v.country,
			v.offset, v.limit)
		assertResult(res, err, v.exp, v.total)
	}
}

func TestGetClientOfferings(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetClientOfferings")
	defer fxt.close()

	other := data.NewTestAccount(data.TestPassword)
	agent := data.NewTestAccount(data.TestPassword)

	var offerings []reform.Record

	testData := []testOfferingData{
		{agent.EthAddr, data.OfferRegistered, data.MsgChPublished,
			"US", false, 11, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferRegistered, data.MsgChPublished,
			"SU", false, 11111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty, "",
			"SU", false, 111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty, "",
			"", true, 111111, fxt.Offering.CurrentSupply},
		{agent.EthAddr, data.OfferRegistered, "",
			"SU", false, 2, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferPoppedUp, data.MsgChPublished,
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

	assertResult := func(res *ui.GetAgentOfferingsResult,
		err error, exp, total int) {
		if res == nil {
			t.Fatal("result is empty")
		}
		if len(res.Items) != exp {
			t.Fatalf("wanted: %v, got: %v", exp, len(res.Items))
		}
		if res.TotalItems != total {
			t.Fatalf("wanted: %v, got: %v", total, res.TotalItems)
		}

		if len(res.Items) > 0 {
			prevBlock := res.Items[0].BlockNumberUpdated

			for _, item := range res.Items {
				current := item.BlockNumberUpdated
				if current <= prevBlock {
					continue
				}

				t.Fatalf("offerings must be ordered by block")
			}
		}
	}

	_, err := handler.GetAgentOfferings("wrong-password", "", "", 0, 0)
	assertMatchErr(ui.ErrAccessDenied, err)

	testArgs := []testGetAgentOfferingsArgs{
		// Test pagination.
		{1, "", "", 0, 1, 2},
		{1, "", "", 1, 0, 2},
		{0, "", "", 2, 0, 2},
		// Test by filters.
		{2, "", "", 0, 0, 2},
		{2, fxt.Product.ID, "", 0, 0, 2},
		{1, "", data.OfferEmpty, 0, 0, 1},
	}

	for _, v := range testArgs {
		res, err := handler.GetAgentOfferings(data.TestPassword,
			v.product, v.offerStatus, v.offset, v.limit)
		assertResult(res, err, v.exp, v.total)
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
	acc1 := createNotUsedAcc(t)
	acc2 := createNotUsedAcc(t)
	defer data.DeleteFromTestDB(t, fxt.DB, acc1, acc2)

	testData := []testOfferingData{
		{fxt.Account.EthAddr, data.OfferRegistered, "", "",
			false, 1, fxt.Offering.CurrentSupply},
		{fxt.Account.EthAddr, data.OfferEmpty, "", "",
			false, 2, fxt.Offering.CurrentSupply},
		{acc1.EthAddr, data.OfferRegistering, "", "",
			false, 2, fxt.Offering.CurrentSupply},
		{acc2.EthAddr, data.OfferRegistered, "", "",
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

	for _, item := range []data.Offering{*offering, *offering} {
		res, err := handler.CreateOffering(data.TestPassword, &item)
		assertMatchErr(nil, err)
		offering2 := &data.Offering{}
		err = db.FindByPrimaryKeyTo(offering2, res)
		assertMatchErr(nil, err)
		defer data.DeleteFromTestDB(t, db, offering2)
	}
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
	handler.SetMockQueue(job.QueueMock(func(method int, tx *reform.TX,
		j2 *data.Job, relatedIDs []string, subID string,
		subFunc job.SubFunc) error {
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
