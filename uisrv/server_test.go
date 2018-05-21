// +build !noagentuisrvtest

package uisrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/eth/truffle"
	"github.com/privatix/dappctrl/util"
)

// used throughout all tests in the package.
var (
	testServer         *Server
	testTruffleAPI     truffle.API
	testEthereumClient *ethclient.Client

	testPassword = "test-password"
)

type testConfig struct {
	ServerStartupDelay uint // In milliseconds.
}

func TestMain(m *testing.M) {
	var conf struct {
		AgentServer     *Config
		AgentServerTest *testConfig
		DB              *data.DBConfig
		Log             *util.LogConfig
		Eth             struct {
			GethURL       string
			TruffleAPIURL string
		}
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	conf.AgentServerTest = &testConfig{}
	util.ReadTestConfig(&conf)
	logger := util.NewTestLogger(conf.Log)
	db := data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)

	var ptc *contract.PrivatixTokenContract
	var psc *contract.PrivatixServiceContract

	if conf.Eth.TruffleAPIURL != "" {
		conn, err := ethclient.Dial(conf.Eth.GethURL)
		if err != nil {
			panic(err)
		}
		testEthereumClient = conn

		testTruffleAPI = truffle.API(conf.Eth.TruffleAPIURL)

		ptcAddr := common.HexToAddress(testTruffleAPI.FetchPTCAddress())
		ptc, err = contract.NewPrivatixTokenContract(ptcAddr, conn)
		if err != nil {
			panic(err)
		}

		pscAddr := common.HexToAddress(testTruffleAPI.FetchPSCAddress())
		psc, err = contract.NewPrivatixServiceContract(pscAddr, conn)
		if err != nil {
			panic(err)
		}
	}
	pwdStorage := new(data.PWDStorage)
	testServer = NewServer(conf.AgentServer, logger, db, testEthereumClient, ptc, psc, pwdStorage)
	testServer.encryptKeyFunc = data.TestEncryptedKey
	testServer.decryptKeyFunc = data.TestToPrivateKey
	go testServer.ListenAndServe()

	time.Sleep(time.Duration(conf.AgentServerTest.ServerStartupDelay) *
		time.Millisecond)

	os.Exit(m.Run())
}

func cleanDB(t *testing.T) {
	data.CleanTestDB(t, testServer.db)
}

func insertItems(t *testing.T, items ...reform.Struct) {
	data.InsertToTestDB(t, testServer.db, items...)
}

func createTestChannel(t *testing.T) *data.Channel {
	agent := data.NewTestUser()
	client := data.NewTestUser()
	product := data.NewTestProduct()
	tplOffer := data.NewTestTemplate(data.TemplateOffer)
	offering := data.NewTestOffering(agent.EthAddr, product.ID, tplOffer.ID)
	ch := data.NewTestChannel(
		agent.EthAddr,
		client.EthAddr,
		offering.ID,
		0,
		1,
		data.ChannelActive)
	insertItems(t,
		agent,
		client,
		product,
		tplOffer,
		offering,
		ch)
	return ch
}

func genEthAddr(t *testing.T) string {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return data.FromBytes(
		crypto.PubkeyToAddress(key.PublicKey).Bytes())
}

func sendPayload(t *testing.T,
	method, path string,
	payload interface{},
) *http.Response {
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal("failed to marshal: ", err)
	}
	url := fmt.Sprintf("http://%s%s", testServer.conf.Addr, path)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal("failed to create a request: ", err)
	}
	req.SetBasicAuth("", testPassword)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal("failed to perform request: ", err)
	}
	return res
}

func getResources(t *testing.T,
	path string,
	params map[string]string,
) *http.Response {
	values := []string{}
	for k, v := range params {
		if v != "" {
			values = append(values, fmt.Sprintf("%s=%s", k, v))
		}
	}
	url := fmt.Sprintf("http://:%s@%s/%s?%s", testPassword,
		testServer.conf.Addr, path, strings.Join(values, "&"))
	res, err := http.Get(url)
	if err != nil {
		t.Fatal("failed to get: ", err)
	}
	return res
}

func testGetResources(t *testing.T, res *http.Response, exp int) {
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to get products: ", res.StatusCode)
	}
	resData := []map[string]interface{}{}
	json.NewDecoder(res.Body).Decode(&resData)
	if exp != len(resData) {
		t.Fatalf("expected %d items, got: %d", exp, len(resData))
	}
}

func setTestUserCredentials(t *testing.T) {
	hash, err := hashPassword("test-salt", testPassword)
	if err != nil {
		t.Fatal("failed to hash password: ", err)
	}
	insertItems(t, &data.Setting{
		Key:   passwordKey,
		Value: string(hash),
		Name:  "password",
	}, &data.Setting{
		Key:   saltKey,
		Value: "test-salt",
		Name:  "salt",
	})
}
