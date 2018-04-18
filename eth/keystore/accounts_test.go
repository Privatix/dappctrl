package keystore

import (
	"testing"
	"os"
	"path/filepath"
	"path"
	"math/big"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/privatix/dappctrl/eth/lib/tests"
	"github.com/privatix/dappctrl/eth/contract"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
	"github.com/privatix/dappctrl/eth/lib"
)

var (
	keystorePath = "/tmp/keystore/"
	pKey         = "adfd7c50906423a1bc64f9927f5c05f2f8cdd815ffed78534ff2a1f15cbcae19"
	password     = "test password"

	// Test sets of dummy data.
	addr1       = [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	addr2       = [20]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	b32Zero     = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	b32Full     = [32]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
)

func TestManagerCreation(t *testing.T) {
	removeContents(keystorePath)

	if man := NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath}); man == nil {
		t.Fatal("Accounts manager should be created, but wasn't.")
	}

	// Even if path wasn't transferred - manager must be created.
	if man := NewAccountsManager(&AccountsManagerConf{KeystorePath:""}); man == nil {
		t.Fatal("Accounts manager should be created, but wasn't.")
	}
}

func TestKeySettingFromHex(t *testing.T) {
	removeContents(keystorePath)

	man := NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath})
	if _, err := man.SetPrivateKeyFromHex(pKey, password); err != nil {
		t.Fatal(err)
	}
}

func TestKeySettingFromJSON(t *testing.T) {
	removeContents(keystorePath)

	// Creating json file
	man := NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath})
	account, err :=  man.SetPrivateKeyFromHex(pKey, password)
	if err != nil {
		t.Fatal(err)
	}

	// Moving created pKey into separate dir
	jsonKeyFilePath := path.Join(os.TempDir(), "pKey.json")
	err = os.Rename(account.URL.Path, jsonKeyFilePath)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to use generated key as JSON
	man =  NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath})
	_, err = man.SetPrivateKeyFromJSON(jsonKeyFilePath, password)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPasswordUpdating(t *testing.T) {
	removeContents(keystorePath)
	man := NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath})
	_, err :=  man.SetPrivateKeyFromHex(pKey, password)
	if err != nil {
		t.Fatal(err)
	}

	man.UpdatePassword(password, "other password")

	// Checking if transactor with new password would be created.
	_, err = man.Transactor("other password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestContractCalling(t *testing.T) {
	removeContents(keystorePath)

	// Fetching contract address and key
	geth := tests.GethEthereumConfig().Geth
	conn, err := ethclient.Dial(geth.Interface())
	failOnErr(t, err, "Failed to connect to the EthereumConf client")

	contractAddress, err :=  lib.NewAddress(fetchPSCAddress())
	failOnErr(t, err, "Failed to parse received contract address")


	psc, err := contract.NewPrivatixServiceContract(contractAddress.Bytes(), conn)
	failOnErr(t, err, "Failed to connect to the EthereumConf client")

	man := NewAccountsManager(&AccountsManagerConf{KeystorePath: keystorePath})
	_, err =  man.SetPrivateKeyFromHex(fetchTestPrivateKey(), password)
	if err != nil {
		t.Fatal(err)
	}

	auth, err := man.Transactor(password)
	if err != nil {
		t.Fatal(err)
	}

	_, err = psc.ThrowEventLogChannelCreated(auth, addr1, addr2, b32Zero, big.NewInt(0), b32Full)
	failOnErr(t, err, "Failed to call ThrowEventLogChannelCreated")
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func failOnErr(t *testing.T, err error, args ...interface{}) {
	if err != nil {
		t.Fatal(args, " / Error details: ", err)
	}
}

func fetchPSCAddress() string {
	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getPSC")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch PSC address. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make(map[string]interface{})
	json.Unmarshal(body, &data)

	return data["contract"].(map[string]interface{})["address"].(string)
}

func fetchTestPrivateKey() string {
	truffleAPI := tests.GethEthereumConfig().TruffleAPI
	response, err := http.Get(truffleAPI.Interface() + "/getKeys")
	if err != nil || response.StatusCode != 200 {
		log.Fatal("Can't fetch private key. It seems that test environment is broken.")
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal("Can't read response body. It seems that test environment is broken.")
	}

	data := make([]interface{}, 0, 0)
	json.Unmarshal(body, &data)

	return data[0].(map[string]interface{})["privateKey"].(string)
}
