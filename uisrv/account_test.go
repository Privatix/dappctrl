// +build !noethtest

package uisrv

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/truffle"
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
		amount      uint64
	}{
		// Wrong id.
		{
			id:          "wrong-id",
			action:      accountDelete,
			destination: data.ContractPSC,
			amount:      1,
		},
		// Wrong action.
		{
			id:          acc.ID,
			action:      "wrong-action",
			destination: data.ContractPTC,
			amount:      1,
		},
		// Wrong destination.
		{
			id:          acc.ID,
			action:      accountTransfer,
			destination: "",
			amount:      1,
		},
		// Wrong amount.
		{
			id:          acc.ID,
			action:      accountTransfer,
			destination: data.ContractPSC,
			amount:      0,
		},
		// Not enough balance.
		{
			id:          acc.ID,
			action:      accountTransfer,
			destination: data.ContractPSC,
			amount:      acc.PSCBalance + 1,
		},
	}

	// Test request parameters validation.
	for _, testCase := range testCases {
		res := sendAccountBalanceAction(
			t,
			testCase.id,
			testCase.action,
			testCase.destination,
			testCase.amount,
		)
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("got: %d for: %+v", res.StatusCode, testCase)
		}
	}

	// TODO:
	// transfer ptc job created
	// delete ptc job created
	// transfer psc job created
	// delete psc job created.
}

func sendAccountBalanceAction(t *testing.T,
	id, action, destination string, amount uint64) *http.Response {
	path := fmt.Sprint(accountsPath, id, "/status")
	payload := &accountBalancePayload{
		Action:      action,
		Amount:      amount,
		Destination: destination,
	}
	return sendPayload(t, http.MethodPut, path, payload)
}

func getTestAccountPayload(testAcc *truffle.TestAccount) *accountCreatePayload {
	payload := &accountCreatePayload{}

	payload.PrivateKey = data.FromBytes(crypto.FromECDSA(testAcc.PrivateKey))

	payload.IsDefault = true
	payload.InUse = true
	payload.Name = "Test account"

	return payload
}

func getTestAccountKeyStorePayload(testAcc *truffle.TestAccount) *accountCreatePayload {
	payload := &accountCreatePayload{}

	pkB64, _ := data.EncryptedKey(testAcc.PrivateKey, payload.JSONKeyStorePassword)
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
	testAcc *truffle.TestAccount,
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

	pubB := crypto.FromECDSAPub(&testAcc.PrivateKey.PublicKey)

	if created.PublicKey != data.FromBytes(pubB) {
		t.Fatal("wrong public key stored")
	}

	addr := crypto.PubkeyToAddress(testAcc.PrivateKey.PublicKey)

	if created.EthAddr != data.FromBytes(addr.Bytes()) {
		t.Fatal("wrong eth addr stored")
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(testServer.conf.EthCallTimeout)*time.Second)
	defer cancel()
	balance, err := testEthereumClient.BalanceAt(ctx, testAcc.Address, nil)
	if err != nil {
		t.Fatal(err)
	}

	if strings.TrimSpace(string(created.EthBalance)) != data.FromBytes(balance.Bytes()) {
		t.Fatal("wrong eth balance stored")
	}

	pscBalance, err := testServer.psc.BalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		t.Fatal(err)
	}

	if created.PSCBalance != pscBalance.Uint64() {
		t.Fatal("wrong psc balance stored")
	}

	ptcBalance, err := testServer.ptc.BalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		t.Fatal(err)
	}

	if created.PTCBalance != ptcBalance.Uint64() {
		t.Logf("got: %d, expected: %d", created.PTCBalance, ptcBalance.Uint64())
		t.Fatal("wrong ptc balance stored")
	}
}

func testCreateAccount(t *testing.T, useRawJSONPayload bool) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	testAccounts, err := testTruffleAPI.GetTestAccounts()
	if err != nil {
		t.Fatal(err)
	}

	testAcc := testAccounts[0]
	var payload *accountCreatePayload
	if useRawJSONPayload {
		payload = getTestAccountKeyStorePayload(&testAcc)
	} else {
		payload = getTestAccountPayload(&testAcc)
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

	testAccountFields(t, &testAcc, payload, created)
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
