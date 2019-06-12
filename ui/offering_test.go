package ui_test

import (
	"encoding/json"
	"errors"
	"testing"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type testOfferingData struct {
	agent              data.HexString
	status             string
	country            string
	isLocal            bool
	blockNumberUpdated uint64
	currentSupply      uint16
}

type testGetAgentOfferingsArgs struct {
	exp           int
	product       string
	offerStatuses []string
	offset        uint
	limit         uint
	total         int
}

type testGetClientOfferingsArgs struct {
	exp          int
	agent        data.HexString
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

type offeringsFilterParamsData struct {
	country    string
	status     string
	setupPrice uint64
	unitPrice  uint64
	minUnits   uint64
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

	minDeposit := data.ComputePrice(fxt.Offering, fxt.Offering.MinUnits)

	_, err := handler.AcceptOffering("wrong-token", fxt.UserAcc.EthAddr,
		fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.AcceptOffering(testToken.v,
		data.HexString(util.NewUUID()), fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(ui.ErrAccountNotFound, err)

	_, err = handler.AcceptOffering(testToken.v, fxt.UserAcc.EthAddr,
		util.NewUUID(), minDeposit, 12345)
	assertErrEqual(ui.ErrOfferingNotFound, err)

	testSOMCClient.Err = errors.New("test error")
	_, err = handler.AcceptOffering(testToken.v, fxt.UserAcc.EthAddr,
		fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(ui.ErrSOMCIsNotAvailable, err)

	testSOMCClient.Err = nil
	_, err = handler.AcceptOffering(testToken.v, fxt.UserAcc.EthAddr,
		fxt.Offering.ID, minDeposit-1, 12345)
	assertErrEqual(ui.ErrDepositTooSmall, err)

	res, err := handler.AcceptOffering(testToken.v, fxt.UserAcc.EthAddr,
		fxt.Offering.ID, minDeposit, 12345)
	assertErrEqual(nil, err)

	if res == nil || j == nil || j.RelatedType != data.JobChannel ||
		j.RelatedID != *res ||
		j.Type != data.JobClientPreChannelCreate {
		t.Fatalf("wrong result data")
	}
}

func createTestOffering(fxt *fixture, agent data.HexString,
	status, country string, isLocal bool,
	blockNumberUpdated uint64, currentSupply uint16) *data.Offering {
	offering := data.NewTestOffering(
		agent, fxt.Product.ID, fxt.TemplateOffer.ID)
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

func TestGetClientOfferings(t *testing.T) {
	fxt, assertErrEqual := newTest(t, "GetClientOfferings")
	defer fxt.close()

	other := data.NewTestAccount(data.TestPassword)
	agent := data.NewTestAccount(data.TestPassword)

	var offerings []reform.Record

	testData := []testOfferingData{
		{agent.EthAddr, data.OfferRegistered,
			"US", false, 11, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferRegistered,
			"SU", false, 11111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty,
			"SU", false, 111, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferEmpty,
			"", true, 111111, fxt.Offering.CurrentSupply},
		{agent.EthAddr, data.OfferRegistered,
			"SU", false, 2, fxt.Offering.CurrentSupply},
		{other.EthAddr, data.OfferPoppedUp,
			"US", false, 222, 0},
	}

	for _, v := range testData {
		offering := createTestOffering(fxt,
			v.agent, v.status, v.country,
			v.isLocal, v.blockNumberUpdated, v.currentSupply)
		offerings = append(offerings, offering)
	}

	for _, offering := range offerings {
		data.InsertToTestDB(t, db, offering)
	}

	defer data.DeleteFromTestDB(t, db, offerings...)

	assertResult := func(res *ui.GetClientOfferingsResult,
		err error, exp, total int) {
		assertErrEqual(nil, err)
		if res == nil {
			t.Fatal("result is empty")
		}
		if len(res.Items) != exp {
			t.Fatalf("wanted items in result: %v, got: %v", exp, len(res.Items))
		}
		if res.TotalItems != total {
			t.Fatalf("wanted total items: %v, got: %v", total, res.TotalItems)
		}
	}

	lowPrice := fxt.Offering.UnitPrice - 10
	price := fxt.Offering.UnitPrice
	highPrice := fxt.Offering.UnitPrice + 10

	_, err := handler.GetClientOfferings(
		"wrong-token", "", 0, 0, nil, 0, 0)
	assertErrEqual(ui.ErrAccessDenied, err)

	_, err = handler.GetClientOfferings(
		testToken.v, "", highPrice, lowPrice, nil, 0, 0)
	assertErrEqual(ui.ErrBadUnitPriceRange, err)

	testArgs := []testGetClientOfferingsArgs{
		// Test pagination.
		{1, "", 0, 0, nil, 0, 1, 4},
		{2, "", 0, 0, nil, 1, 2, 4},
		{0, "", 0, 0, nil, 4, 3, 4},
		// // Test by filters.
		{4, "", 0, 0, nil, 0, 0, 4},
		{0, "", 0, lowPrice, nil, 0, 0, 0},
		{0, "", highPrice, 0, nil, 0, 0, 0},
		{4, "", lowPrice, price, nil, 0, 0, 4},
		{4, "", price, highPrice, nil, 0, 0, 4},
		{4, "", price, price, nil, 0, 0, 4},
		{1, "", 0, 0, []string{"US"}, 0, 0, 1},
		{2, "", 0, 0, []string{"SU"}, 0, 0, 2},
		{3, "", 0, 0, []string{"SU", "US"}, 0, 0, 3},
		{0, data.NewTestAccount(data.TestPassword).EthAddr, 0, 0, nil, 0, 0, 0},
		{2, agent.EthAddr, 0, 0, nil, 0, 0, 2},
	}

	for _, v := range testArgs {
		res, err := handler.GetClientOfferings(testToken.v,
			v.agent, v.minUnitPrice, v.maxUnitPrice, v.country,
			v.offset, v.limit)
		assertResult(res, err, v.exp, v.total)
	}
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

	_, err := handler.GetAgentOfferings("wrong-token", "", []string{}, 0, 0)
	assertMatchErr(ui.ErrAccessDenied, err)

	testArgs := []testGetAgentOfferingsArgs{
		// Test pagination.
		{1, "", []string{}, 0, 1, 2},
		{1, "", []string{}, 1, 0, 2},
		{0, "", []string{}, 2, 0, 2},
		// Test by filters.
		{2, "", []string{}, 0, 0, 2},
		{2, fxt.Product.ID, []string{}, 0, 0, 2},
		{1, "", []string{data.OfferEmpty}, 0, 0, 1},
	}

	for _, v := range testArgs {
		res, err := handler.GetAgentOfferings(testToken.v,
			v.product, v.offerStatuses, v.offset, v.limit)
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
		{fxt.Account.EthAddr, data.OfferRegistered, "",
			false, 1, fxt.Offering.CurrentSupply},
		{fxt.Account.EthAddr, data.OfferEmpty, "",
			false, 2, fxt.Offering.CurrentSupply},
		{acc1.EthAddr, data.OfferRegistering, "",
			false, 2, fxt.Offering.CurrentSupply},
		{acc2.EthAddr, data.OfferRegistered, "",
			false, 2, fxt.Offering.CurrentSupply},
	}

	for _, v := range testData {
		offering := createTestOffering(fxt,
			v.agent, v.status, v.country,
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

	for i, v := range invalidOfferings {
		_, err := handler.CreateOffering(testToken.v, v)
		if err == nil {
			t.Fatalf("offering %d should not be saved", i)
		}
	}

	offering := data.NewTestOffering(data.HexString(fxt.Account.ID),
		fxt.Product.ID, fxt.TemplateOffer.ID)

	_, err := handler.CreateOffering("wrong-token", offering)
	assertMatchErr(ui.ErrAccessDenied, err)

	for _, item := range []data.Offering{*offering, *offering} {
		res, err := handler.CreateOffering(testToken.v, &item)
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

	err := handler.UpdateOffering("wrong-token", fxt.Offering)
	assertMatchErr(ui.ErrAccessDenied, err)

	newOffering := data.NewTestOffering(data.HexString(fxt.Account.ID),
		fxt.Product.ID, fxt.TemplateOffer.ID)

	err = handler.UpdateOffering(testToken.v, newOffering)
	assertMatchErr(ui.ErrOfferingNotFound, err)

	err = handler.UpdateOffering(testToken.v, fxt.Offering)
	assertMatchErr(nil, err)

	savedOffering := &data.Offering{}
	err = db.FindByPrimaryKeyTo(savedOffering, fxt.Offering.ID)
	if err != nil {
		t.Fatal(err)
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
			testToken.v, fxt.Offering.ID, action, 100)
		assertMatchErr(nil, err)

		if j == nil || j.Type != jobType ||
			j.RelatedID != fxt.Offering.ID {
			t.Fatal("expected job not created")
		}
	}
}

func TestGetClientOfferingsFilterParams(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetClientOfferingsFilterParams")
	defer fxt.close()

	agent := data.NewTestAccount(data.TestPassword)

	c1 := "AB"
	c2 := "CD"

	testData := []*offeringsFilterParamsData{
		{c1, data.OfferRegistered, 1, 1, 1},
		{c1, data.OfferRegistered, 2, 2, 2},
		{c2, data.OfferRegistered, 10, 10, 10},
		// Ignored offerings.
		{"YY", data.OfferEmpty, 20, 20, 20},
		{"", data.OfferRegistered, 20, 20, 20},
	}

	var offerings []*data.Offering

	for _, v := range testData {
		offering := data.NewTestOffering(data.HexString(agent.ID),
			fxt.Product.ID, fxt.TemplateOffer.ID)
		offering.Country = v.country
		offering.Status = v.status
		offering.SetupPrice = v.setupPrice
		offering.UnitPrice = v.unitPrice
		offering.MinUnits = v.minUnits
		offerings = append(offerings, offering)
	}

	min := offerings[0].UnitPrice
	max := offerings[2].UnitPrice

	for _, v := range offerings {
		data.InsertToTestDB(t, db, v)
		defer data.DeleteFromTestDB(t, db, v)
	}

	_, err := handler.GetClientOfferingsFilterParams("wrong-token")
	assertMatchErr(ui.ErrAccessDenied, err)

	res, err := handler.GetClientOfferingsFilterParams(testToken.v)
	assertMatchErr(nil, err)

	if len(res.Countries) != 2 {
		t.Fatalf("wanted: %v, got: %v", 2, res.Countries)
	}

	if res.Countries[0] != c1 {
		t.Fatalf("wanted: %v, got: %v", c1, res.Countries[0])
	}

	if res.Countries[1] != c2 {
		t.Fatalf("wanted: %v, got: %v", c2, res.Countries[1])
	}

	if res.MinPrice != min {
		t.Fatalf("wanted: %v, got: %v", min, res.MinPrice)
	}

	if res.MaxPrice != max {
		t.Fatalf("wanted: %v, got: %v", max, res.MaxPrice)
	}
}

func TestPingOfferings(t *testing.T) {
	fxt, assertErrorEquals := newTest(t, "PingOfferings")
	defer fxt.close()

	_, err := handler.PingOfferings("wrong-token", []string{"sdfs"})
	assertErrorEquals(ui.ErrAccessDenied, err)

	_, err = handler.PingOfferings(testToken.v, []string{util.NewUUID()})
	assertErrorEquals(ui.ErrOfferingNotFound, err)

	offering := *fxt.Offering
	offering.ID = util.NewUUID()
	offering.Hash = data.HexString("sdfsdf")
	data.InsertToTestDB(t, fxt.DB, &offering)
	defer data.DeleteFromTestDB(t, fxt.DB, &offering)
	ret, err := handler.PingOfferings(testToken.v, []string{fxt.Offering.ID, offering.ID})
	assertErrorEquals(nil, err)
	if !ret[fxt.Offering.ID] || !ret[offering.ID] {
		t.Fatalf("wrong ping result: got %v", ret)
	}
	fxt.DB.Reload(fxt.Offering)
	if fxt.Offering.SOMCSuccessPing == nil {
		t.Fatalf("somc success ping time not recorded")
	}

	testSOMCClient.Err = errors.New("test error")
	ret, err = handler.PingOfferings(testToken.v, []string{fxt.Offering.ID})
	assertErrorEquals(nil, err)
	if ret[fxt.Offering.ID] {
		t.Fatalf("wrong ping result, got %v", ret)
	}
}
