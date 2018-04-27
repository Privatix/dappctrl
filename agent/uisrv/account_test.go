// +build !noethtest

package uisrv

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/truffle"
)

func TestUpdateAccountCheckAvailableBalance(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	acc := data.NewTestAccount()
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

	pkB, _ := hex.DecodeString(testAcc.PrivateKey)
	payload.PrivateKey = data.FromBytes(pkB)

	payload.IsDefault = true
	payload.InUse = true
	payload.Name = "Test account"

	return payload
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

	createdPK, err := data.ToPrivateKey(created.PrivateKey, testPassword)
	if err != nil {
		t.Fatal("failed to decrypt created account's private key: ", err)
	}

	if data.FromBytes(crypto.FromECDSA(createdPK)) != payload.PrivateKey {
		t.Fatal("wrong private key stored")
	}

	pkB, err := hex.DecodeString(testAcc.PrivateKey)
	if err != nil {
		t.Fatal("failed to decode priv key: ", err)
	}
	pk, err := crypto.ToECDSA(pkB)
	if err != nil {
		t.Fatal("failed to make ecdsa: ", err)
	}

	pubB := crypto.FromECDSAPub(&pk.PublicKey)

	if created.PublicKey != data.FromBytes(pubB) {
		t.Fatal("wrong public key stored")
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	if created.EthAddr != data.FromBytes(addr.Bytes()) {
		t.Fatal("wrong eth addr stored")
	}

	response, err := testEthereumClient.GetBalance(testAcc.Account, eth.BlockLatest)
	if err != nil {
		t.Fatal(err)
	}

	amount, err := eth.NewUint192(response.Result)
	if err != nil {
		t.Fatal(err, response.Result)
	}
	if strings.TrimSpace(string(created.EthBalance)) != data.FromBytes(amount.ToBigInt().Bytes()) {
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

func TestCreateAccount(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	testAcc := testTruffleAPI.GetTestAccounts()[0]
	payload := getTestAccountPayload(&testAcc)

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

func TestGetAccounts(t *testing.T) {
	defer cleanDB(t)
	setTestUserCredentials(t)

	// Test returns empty all accounts in the system.

	res := getResources(t, accountsPath, nil)
	testGetResources(t, res, 0)

	acc1 := data.NewTestAccount()
	acc2 := data.NewTestAccount()
	insertItems(t, acc1, acc2)

	res = getResources(t, accountsPath, nil)
	testGetResources(t, res, 2)

	// get account by id.
	res = getResources(t, accountsPath, map[string]string{"id": acc1.ID})
	testGetResources(t, res, 1)
}
