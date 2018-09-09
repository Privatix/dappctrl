package ui_test

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
)

type accountCreatePayload struct {
	privateKey           string
	jsonKeyStoreRaw      json.RawMessage
	jsonKeyStorePassword string
	isDefault            bool
	inUse                bool
	name                 string
}

type transferTokensPayload struct {
	account     string
	destination string
	amount      uint64
	gasPrice    uint64
}

func TestExportPrivateKey(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "ExportPrivateKey")
	defer fxt.close()

	_, err := handler.ExportPrivateKey("wrong-password", fxt.Account.ID)
	assertMatchErr(ui.ErrAccessDenied, err)

	expectedBytes := []byte(`{"hello": "world"}`)
	fxt.Account.PrivateKey = data.FromBytes(expectedBytes)

	data.SaveToTestDB(t, db, fxt.Account)

	res, err := handler.ExportPrivateKey(data.TestPassword, fxt.Account.ID)
	assertMatchErr(nil, err)

	if !bytes.Equal(res.PrivateKey, expectedBytes) {
		t.Fatalf("wrong private key exported:"+
			" expected %x got %x", expectedBytes, res.PrivateKey)
	}
}

func TestGetAccounts(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "GetAccounts")
	defer fxt.close()

	expectedAccNumber := 2 // from fixture

	_, err := handler.GetAccounts("wrong-password")
	assertMatchErr(ui.ErrAccessDenied, err)

	res, err := handler.GetAccounts(data.TestPassword)
	assertMatchErr(nil, err)

	if len(res.Accounts) != expectedAccNumber {
		t.Fatalf("expected %d items, got: %d (%s)",
			expectedAccNumber, len(res.Accounts), util.Caller())
	}
}

func newCreatePayload() *accountCreatePayload {
	return &accountCreatePayload{
		isDefault: true,
		inUse:     true,
		name:      "Test account",
	}
}

func privateKeyToJSON(pk *ecdsa.PrivateKey) (key json.RawMessage, pass string) {
	pkB64, _ := data.EncryptedKey(pk, pass)
	jsonBytes, _ := data.ToBytes(pkB64)
	return jsonBytes, pass
}

func fromPrivateKeyToECDSA(t *testing.T, privateKey string) *ecdsa.PrivateKey {
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

func fromJSONKeyStoreRawToECDSA(t *testing.T, jsonKeyStoreRaw json.RawMessage,
	jsonKeyStorePassword string) *ecdsa.PrivateKey {
	key, err := keystore.DecryptKey(
		jsonKeyStoreRaw, jsonKeyStorePassword)
	if err != nil {
		t.Fatal(err)
	}
	return key.PrivateKey
}

func toECDSA(t *testing.T, privateKey, jsonKeyStorePassword string,
	jsonKeyStoreRaw json.RawMessage) *ecdsa.PrivateKey {
	var pk *ecdsa.PrivateKey
	if privateKey != "" {
		pk = fromPrivateKeyToECDSA(t, privateKey)
	} else if len(jsonKeyStoreRaw) != 0 {
		pk = fromJSONKeyStoreRawToECDSA(
			t, jsonKeyStoreRaw, jsonKeyStorePassword)
	}
	return pk
}

func checkTestAccount(t *testing.T, pk *ecdsa.PrivateKey,
	payload *accountCreatePayload, created *data.Account) {
	if created.Name != payload.name {
		t.Fatal("wrong name stored")
	}

	if created.IsDefault != payload.isDefault {
		t.Fatal("wrong is default stored")
	}

	if created.InUse != payload.inUse {
		t.Fatal("wrong in use stored")
	}

	expectedKey := toECDSA(t, payload.privateKey,
		payload.jsonKeyStorePassword, payload.jsonKeyStoreRaw)

	createdKey, err := data.TestToPrivateKey(created.PrivateKey, "")
	if err != nil {
		t.Fatal("failed to decrypt account's private key: ", err)
	}

	equalECDSA := func(a, b *ecdsa.PrivateKey) bool {
		abytes := crypto.FromECDSA(a)
		bbytes := crypto.FromECDSA(b)
		return bytes.Compare(abytes, bbytes) == 0
	}

	if !equalECDSA(expectedKey, createdKey) {
		t.Fatal("wrong private key stored")
	}

	pubB := crypto.FromECDSAPub(&pk.PublicKey)

	if created.PublicKey != data.FromBytes(pubB) {
		t.Fatal("wrong public key stored")
	}
}

func testCreateAccount(t *testing.T, useJSONKey bool) {
	fxt, _ := newTest(t, "CreateAccount")
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

	payload := newCreatePayload()

	pk, _ := crypto.GenerateKey()
	if useJSONKey {
		payload.privateKey = data.FromBytes(crypto.FromECDSA(pk))
	} else {
		k, p := privateKeyToJSON(pk)
		payload.jsonKeyStoreRaw, payload.jsonKeyStorePassword = k, p
	}

	res, err := handler.CreateAccount(data.TestPassword, payload.privateKey,
		payload.jsonKeyStoreRaw, payload.jsonKeyStorePassword,
		payload.name, payload.isDefault, payload.inUse)
	if err != nil {
		t.Fatal(err)
	}

	account := &data.Account{}
	err = db.FindByPrimaryKeyTo(account, res.Account)
	if err != nil {
		t.Fatal(err)
	}
	defer data.DeleteFromTestDB(t, db, account)

	checkTestAccount(t, pk, payload, account)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != account.ID ||
		j.Type != data.JobAccountUpdateBalances ||
		j.CreatedBy != data.JobUser {
		t.Fatalf("expected job not created")
	}
}

func TestCreateAccount(t *testing.T) {
	testCreateAccount(t, true)
	testCreateAccount(t, false)
}

func TestTransferTokens(t *testing.T) {
	fxt, assertMatchErr := newTest(t, "TransferTokens")
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

	res := handler.TransferTokens("wrong-password",
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
		res := handler.TransferTokens(data.TestPassword,
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

	res = handler.TransferTokens(data.TestPassword,
		fxt.Account.ID, data.ContractPTC,
		payload.amount, payload.gasPrice)
	assertMatchErr(nil, res)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != fxt.Account.ID ||
		j.Type != data.JobPreAccountReturnBalance {
		t.Fatalf("wrong result job")
	}
	checkJobDataFields(j.Data, payload.amount, payload.gasPrice)

	res = handler.TransferTokens(data.TestPassword,
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

	res := handler.UpdateBalance("wrong-password", fxt.Account.ID)
	assertMatchErr(ui.ErrAccessDenied, res)

	res = handler.UpdateBalance(data.TestPassword, fxt.Account.ID)
	assertMatchErr(nil, res)

	if j == nil || j.RelatedType != data.JobAccount ||
		j.RelatedID != fxt.Account.ID ||
		j.Type != data.JobAccountUpdateBalances {
		t.Fatalf("wrong result job")
	}
}
