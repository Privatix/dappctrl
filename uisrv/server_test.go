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

	"github.com/AlekSi/pointer"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// used throughout all tests in the package.
var (
	testServer         *Server
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
		Job             *job.Config
		StderrLog       *log.WriterConfig
	}
	conf.DB = data.NewDBConfig()
	conf.AgentServerTest = &testConfig{}
	conf.StderrLog = log.NewWriterConfig()
	util.ReadTestConfig(&conf)
	db := data.NewTestDB(conf.DB)
	defer data.CloseDB(db)

	logger, err := log.NewStderrLogger(conf.StderrLog)
	if err != nil {
		panic(err)
	}

	queue := job.NewQueue(conf.Job, logger, db, nil)

	pwdStorage := new(data.PWDStorage)
	testServer = NewServer(conf.AgentServer, logger, db, "", queue,
		pwdStorage, proc.NewProcessor(proc.NewConfig(), db, queue))
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
	offering.MaxInactiveTimeSec = pointer.ToUint64(10)
	offering.UnitName = "megabytes"
	offering.UnitType = data.UnitScalar
	offering.SetupPrice = 22
	offering.UnitPrice = 11
	ch := data.NewTestChannel(
		agent.EthAddr,
		client.EthAddr,
		offering.ID,
		0,
		10000,
		data.ChannelActive)
	ch.ServiceChangedTime = pointer.ToTime(time.Now())
	insertItems(t,
		agent,
		client,
		tplOffer,
		product,
		offering,
		ch)
	return ch
}

func genEthAddr(t *testing.T) data.HexString {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	return data.HexFromBytes(
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

func testGetResources(t *testing.T, res *http.Response, exp int) []map[string]interface{} {
	if res.StatusCode != http.StatusOK {
		t.Fatal("failed to get resources: ", res.StatusCode)
	}
	ret := []map[string]interface{}{}
	json.NewDecoder(res.Body).Decode(&ret)
	if exp != len(ret) {
		t.Fatalf("expected %d items, got: %d (%s)", exp, len(ret),
			util.Caller())
	}
	return ret
}

func setTestUserCredentials(t *testing.T) func() {
	hash, err := data.HashPassword(testPassword, "test-salt")
	if err != nil {
		t.Fatal("failed to hash password: ", err)
	}
	pwdSetting := &data.Setting{
		Key:   passwordKey,
		Value: string(hash),
		Name:  "password",
	}

	saltSetting := &data.Setting{
		Key:   saltKey,
		Value: "test-salt",
		Name:  "salt",
	}

	data.InsertToTestDB(t, testServer.db, pwdSetting, saltSetting)
	return func() {
		data.DeleteFromTestDB(t, testServer.db, pwdSetting, saltSetting)
	}
}
