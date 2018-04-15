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

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

// used throughout all tests in the package.
var (
	testServer *Server
)

func TestMain(m *testing.M) {
	var conf struct {
		AgentServer *Config
		DB          *data.DBConfig
		Log         *util.LogConfig
	}
	conf.DB = data.NewDBConfig()
	conf.Log = util.NewLogConfig()
	util.ReadTestConfig(&conf)
	logger := util.NewTestLogger(conf.Log)
	db := data.NewTestDB(conf.DB, logger)
	defer data.CloseDB(db)
	testServer = NewServer(conf.AgentServer, logger, db)
	go testServer.ListenAndServe()
	os.Exit(m.Run())
}

func cleanDB() {
	data.CleanDB(testServer.db)
}

func insertItems(items ...reform.Struct) {
	data.InsertItems(testServer.db, items...)
}

func createTestChannel() *data.Channel {
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
	insertItems(
		agent,
		client,
		product,
		tplOffer,
		offering,
		ch)
	return ch
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
	url := fmt.Sprintf("http://%s/%s?%s", testServer.conf.Addr, path, strings.Join(values, "&"))
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
