package ui_test

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type transferTokensPayload struct {
	account     string
	destination string
	amount      uint64
	gasPrice    uint64
}

func TestExportPrivateKey(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "ExportPrivateKey")
	defer fxt.close()

	_, err := handler.ExportPrivateKey("wrong-token", fxt.Account.ID)
	assertMatchErr(ui.ErrAccessDenied, err)

	expectedBytes := []byte(`{"hello": "world"}`)
	fxt.Account.PrivateKey = data.FromBytes(expectedBytes)

	data.SaveToTestDB(t, db, fxt.Account)

	res, err := handler.ExportPrivateKey(testToken.v, fxt.Account.ID)
	assertMatchErr(nil, err)

	if !bytes.Equal(res, expectedBytes) {
		t.Fatalf("wrong private key exported:"+
			" expected %x got %x", expectedBytes, res)
	}
}

func TestGetAccounts(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetAccounts")
	defer fxt.close()

	expectedAccNumber := 2 // from fixture

	_, err := handler.GetAccounts("wrong-token")
	assertMatchErr(ui.ErrAccessDenied, err)

	res, err := handler.GetAccounts(testToken.v)
	assertMatchErr(nil, err)

	if len(res) != expectedAccNumber {
		t.Fatalf("expected %d items, got: %d (%s)",
			expectedAccNumber, len(res), util.Caller())
	}
}

func privateKeyToJSON(pk *ecdsa.PrivateKey) (key json.RawMessage, pass string) {
	pkB64, _ := data.EncryptedKey(pk, pass)
	jsonBytes, _ := data.ToBytes(pkB64)
	return jsonBytes, pass
}

func fromPrivateKeyToECDSA(t *testing.T,
	privateKey data.Base64String) *ecdsa.PrivateKey {
	pkBytes, err := data.ToBytes(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	key, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testAccountFields(t *testing.T,
	expected *ui.AccountParams, created *data.Account) {
	if created.Name != expected.Name {
		t.Fatal("wrong name stored")
	}

	if created.IsDefault != expected.IsDefault {
		t.Fatal("wrong is default stored")
	}

	if created.InUse != expected.InUse {
		t.Fatal("wrong in use stored")
	}
}

func checkGeneratedAccount(t *testing.T,
	expected *ui.AccountParams, created *data.Account) {
	testAccountFields(t, expected, created)

	_, err := data.TestToPrivateKey(created.PrivateKey, "")
	if err != nil {
		t.Fatal("failed to decrypt account's private key: ", err)
	}
}

func testAccount(t *testing.T, expected *ui.AccountParams,
	created *data.Account, key *ecdsa.PrivateKey) {
	testAccountFields(t, expected, created)

	expectedKey := fromPrivateKeyToECDSA(
		t, data.FromBytes(crypto.FromECDSA(key)))

	createdKey, err := data.TestToPrivateKey(created.PrivateKey, "")
	if err != nil {
		t.Fatal("failed to decrypt account's private key: ", err)
	}

	equalECDSA := func(a, b *ecdsa.PrivateKey) bool {
		abytes := crypto.FromECDSA(a)
		bbytes := crypto.FromECDSA(b)
		return bytes.Equal(abytes, bbytes)
	}

	if !equalECDSA(expectedKey, createdKey) {
		t.Fatal("wrong private key stored")
	}

	pubB := crypto.FromECDSAPub(&key.PublicKey)

	if created.PublicKey != data.FromBytes(pubB) {
		t.Fatal("wrong public key stored")
	}
}

func testImportAccount(t *testing.T, expID *string,
	params *ui.AccountParams, expPK *ecdsa.PrivateKey, expJob *data.Job) {
	account2 := &data.Account{}
	err := data.FindByPrimaryKeyTo(db.Querier, account2, expID)
	if err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, account2)

	testAccount(t, params, account2, expPK)

	if expJob == nil || expJob.RelatedType != data.JobAccount ||
		expJob.Type != data.JobAccountUpdateBalances ||
		expJob.CreatedBy != data.JobUser {
		t.Fatalf("expected job not created")
	}
}

func TestGenerateAccount(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GenerateAccount")
	defer fxt.close()

	params := &ui.AccountParams{}
	params.Name = util.NewUUID()[:30]

	res, err := handler.GenerateAccount(testToken.v, params)
	assertMatchErr(nil, err)

	account := &data.Account{}
	err = data.FindByPrimaryKeyTo(db.Querier, account, res)
	if err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, account)

	checkGeneratedAccount(t, params, account)
}

func TestImportAccountFromHex(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "ImportAccountFromHex")
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

	pk, _ := crypto.GenerateKey()

	params := &ui.AccountParamsWithHexKey{}
	params.Name = util.NewUUID()[:30]
	params.PrivateKeyHex = data.HexFromBytes(crypto.FromECDSA(pk))

	res, err := handler.ImportAccountFromHex(testToken.v, params)
	assertMatchErr(nil, err)

	testImportAccount(t, res, &params.AccountParams, pk, j)
}

func TestImportAccountFromJSON(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "ImportAccountFromJSON")
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

	pk, _ := crypto.GenerateKey()
	key, pass := privateKeyToJSON(pk)
	params := &ui.AccountParams{}
	params.Name = util.NewUUID()[:30]

	res, err := handler.ImportAccountFromJSON(
		testToken.v, params, key, pass)
	assertMatchErr(nil, err)

	testImportAccount(t, res, params, pk, j)
}

func TestTransferTokens(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "TransferTokens")
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

	res := handler.TransferTokens("wrong-token",
		fxt.Account.ID, data.ContractPSC, 1, 1)
	assertMatchErr(ui.ErrAccessDenied, res)

	testCases := []*transferTokensPayload{
		// Wrong account.
		{
			account:     util.NewUUID(),
			destination: data.ContractPSC,
			amount:      1,
		},
		// Wrong destination.
		{
			account:     fxt.Account.ID,
			destination: "",
			amount:      1,
		},
		// Wrong amount.
		{
			account:     fxt.Account.ID,
			destination: data.ContractPSC,
			amount:      0,
		},
	}

	for _, testCase := range testCases {
		res := handler.TransferTokens(testToken.v,
			testCase.account, testCase.destination,
			testCase.amount, testCase.gasPrice)
		if res == nil {
			t.Fatal("error must be not nil")
		}
	}

	payload := &transferTokensPayload{
		account:  fxt.Account.ID,
		amount:   1,
		gasPrice: 20,
	}

	checkJobDataFields := func(jobRawData []byte, amount, gasPrice uint64) {
		var jobData *data.JobBalanceData

		if err := json.Unmarshal(jobRawData, &jobData); err != nil {
			t.Fatal(err)
		}

		if jobData.Amount != amount || jobData.GasPrice != gasPrice {
			t.Fatalf("wrong job data fields")
		}
	}

	res = handler.TransferTokens(testToken.v,
		fxt.Account.ID, data.ContractPTC,
		payload.amount, payload.gasPrice)
	assertMatchErr(nil, res)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != fxt.Account.ID ||
		j.Type != data.JobPreAccountReturnBalance {
		t.Fatalf("wrong result job")
	}
	checkJobDataFields(j.Data, payload.amount, payload.gasPrice)

	res = handler.TransferTokens(testToken.v,
		fxt.Account.ID, data.ContractPSC,
		payload.amount, payload.gasPrice)
	assertMatchErr(nil, res)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != fxt.Account.ID ||
		j.Type != data.JobPreAccountAddBalanceApprove {
		t.Fatalf("wrong result job")
	}
	checkJobDataFields(j.Data, payload.amount, payload.gasPrice)
}

func TestUpdateBalance(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "UpdateBalance")
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

	res := handler.UpdateBalance("wrong-token", fxt.Account.ID)
	assertMatchErr(ui.ErrAccessDenied, res)

	res = handler.UpdateBalance(testToken.v, fxt.Account.ID)
	assertMatchErr(nil, res)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != fxt.Account.ID ||
		j.Type != data.JobAccountUpdateBalances {
		t.Fatalf("wrong result job")
	}
}

func TestUpdateAccount(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "UpdateAccount")
	defer fxt.close()

	res := handler.UpdateAccount(
		"wrong-token", fxt.Account.ID, "", false, false)
	assertMatchErr(ui.ErrAccessDenied, res)

	acc := data.NewTestAccount(data.TestPassword)
	fxt.DB.Insert(acc)
	defer fxt.DB.Delete(acc)

	for _, td := range []struct {
		acc       *data.Account
		name      string
		isDefault bool
		inUse     bool
	}{
		{
			acc:       fxt.Account,
			name:      "newname",
			isDefault: true,
			inUse:     true,
		},
		{
			acc:       acc,
			name:      acc.Name,
			isDefault: true,
			inUse:     true,
		},
	} {
		res = handler.UpdateAccount(testToken.v,
			td.acc.ID, td.name, td.isDefault, td.inUse)
		assertMatchErr(nil, res)

		db.Reload(td.acc)

		if td.acc.Name != td.name {
			t.Fatalf("expected account name: %s, got: %s",
				td.name, td.acc.Name)
		}

		if td.acc.InUse != td.inUse ||
			td.acc.IsDefault != td.isDefault {
			t.Fatal("failed to update account status")
		}
	}

	ret, _ := fxt.DB.SelectAllFrom(data.AccountTable, "WHERE is_default")
	if len(ret) != 1 {
		t.Fatal("only 1 account allowed to be default")
	}
}
