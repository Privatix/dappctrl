package uisrv

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
)

func TestUpdateAccountCheckAvailableBalance(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	acc := data.NewTestAccount(testPassword)
	insertItems(t, acc)

	testCases := []struct {
		id          string
		action      string
		destination string
		amount      uint
	}{
		// Wrong destination.
		{
			id:          acc.ID,
			destination: "",
			amount:      1,
		},
		// Wrong amount.
		{
			id:          acc.ID,
			destination: data.ContractPSC,
			amount:      0,
		},
	}

	// Test request parameters validation.
	for _, testCase := range testCases {
		res := sendAccountBalanceAction(
			t,
			testCase.id,
			testCase.destination,
			testCase.amount,
		)
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("got: %d for: %+v", res.StatusCode, testCase)
		}
	}

	// TODO:
	// transfer ptc job created
	res := sendAccountBalanceAction(
		t,
		acc.ID,
		data.ContractPTC,
		1,
	)
	if res.StatusCode != http.StatusOK {
		t.Fatal("got: ", res.Status)
	}
	data.FindInTestDB(t, testServer.db, &data.Job{}, "type",
		data.JobPreAccountReturnBalance)
	// transfer psc job created
	res = sendAccountBalanceAction(
		t,
		acc.ID,
		data.ContractPSC,
		1,
	)
	if res.StatusCode != http.StatusOK {
		t.Fatal("got: ", res.Status)
	}
	data.FindInTestDB(t, testServer.db, &data.Job{}, "type",
		data.JobPreAccountAddBalanceApprove)
}

func sendAccountBalanceAction(t *testing.T,
	id, destination string, amount uint) *http.Response {
	path := fmt.Sprint(accountsPath, id, "/status")
	payload := &accountBalancePayload{
		Amount:      amount,
		Destination: destination,
	}
	return sendPayload(t, http.MethodPut, path, payload)
}

func getTestAccountPayload(testkey *ecdsa.PrivateKey) *accountCreatePayload {
	payload := &accountCreatePayload{}

	payload.PrivateKey = data.FromBytes(crypto.FromECDSA(testkey))

	payload.IsDefault = true
	payload.InUse = true
	payload.Name = "Test account"

	return payload
}

func getTestAccountKeyStorePayload(testkey *ecdsa.PrivateKey) *accountCreatePayload {
	payload := &accountCreatePayload{}

	pkB64, _ := data.EncryptedKey(testkey, payload.JSONKeyStorePassword)
	jsonBytes, _ := data.ToBytes(pkB64)
	payload.JSONKeyStoreRaw = string(jsonBytes)

	payload.IsDefault = true
	payload.InUse = true
	payload.Name = "Test account"

	return payload
}

func equalECDSA(a, b *ecdsa.PrivateKey) bool {
	abytes := crypto.FromECDSA(a)
	bbytes := crypto.FromECDSA(b)
	return bytes.Compare(abytes, bbytes) == 0
}

func testAccountFields(
	t *testing.T,
	testkey *ecdsa.PrivateKey,
	payload *accountCreatePayload,
	created *data.Account) {

	if created.Name != payload.Name {
		t.Fatal("wrong name stored")
	}

	if created.IsDefault != payload.IsDefault {
		t.Fatal("wrong is default stored")
	}

	if created.InUse != payload.InUse {
		t.Fatal("wrong in use stored")
	}

	payloadKey, err := payload.toECDSA()
	if err != nil {
		t.Fatalf("could not extract private key from payload: %v", err)
	}

	createdKey, err := data.TestToPrivateKey(created.PrivateKey, testPassword)
	if err != nil {
		t.Fatal("failed to decrypt created account's private key: ", err)
	}

	if !equalECDSA(payloadKey, createdKey) {
		t.Fatal("wrong private key stored")
	}

	pubB := crypto.FromECDSAPub(&testkey.PublicKey)

	if created.PublicKey != data.FromBytes(pubB) {
		t.Fatal("wrong public key stored")
	}
}

func testCreateAccount(t *testing.T, useRawJSONPayload bool) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	testkey, _ := crypto.GenerateKey()
	var payload *accountCreatePayload
	if useRawJSONPayload {
		payload = getTestAccountKeyStorePayload(testkey)
	} else {
		payload = getTestAccountPayload(testkey)
	}

	res := sendPayload(t, http.MethodPost, accountsPath, payload)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("response: %d, wanted: %d", res.StatusCode, http.StatusCreated)
	}

	reply := &replyEntity{}
	json.NewDecoder(res.Body).Decode(reply)
	defer res.Body.Close()

	created := &data.Account{}
	if err := testServer.db.FindByPrimaryKeyTo(created, reply.ID); err != nil {
		t.Fatal("failed to retrieve created account: ", err)
	}

	testAccountFields(t, testkey, payload, created)
}

func TestCreateAccount(t *testing.T) {
	testCreateAccount(t, false)
	testCreateAccount(t, true)
}

func TestExportAccountPrivateKey(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	acc := data.NewTestAccount(testPassword)
	expectedBytes := []byte(`{"hello": "world"}`)
	acc.PrivateKey = data.FromBytes(expectedBytes)
	insertItems(t, acc)

	res := sendPayload(t, http.MethodGet, accountsPath+acc.ID+"/pkey", nil)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("response: %d, wanted: %d", res.StatusCode, http.StatusOK)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if !bytes.Equal(body, expectedBytes) {
		t.Fatalf("wrong pkey exported: expected %x got %x", expectedBytes, body)
	}
}

func TestGetAccounts(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Test returns empty all accounts in the system.

	res := getResources(t, accountsPath, nil)
	testGetResources(t, res, 0)

	acc1 := data.NewTestAccount(testPassword)
	acc2 := data.NewTestAccount(testPassword)
	insertItems(t, acc1, acc2)

	res = getResources(t, accountsPath, nil)
	testGetResources(t, res, 2)

	// get account by id.
	res = getResources(t, accountsPath, map[string]string{"id": acc1.ID})
	testGetResources(t, res, 1)
}
