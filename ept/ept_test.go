package ept

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		EptTest *eptTestConfig
	}

	testEMT *EndpointMessageTemplate
)

const (
	confFileName = "server.conf"
	certFileName = "ca.crt"

	tempPrefix = "eptTest"

	filePerm os.FileMode = 0666

	// In the future, we can change the way files are generated
	validConf   = openVpnConfServerFullExample
	validCert   = certValidExample
	invalidConf = openVpnConfServerFakeCertificate
	invalidCert = certInvalidExample

	errPars = "incorrect parsing test"
	errGen  = "incorrect generate message test"
)

type eptTestConfig struct {
	ExportConfigKeys []string
	ValidHash        []string
	InvalidHash      []string
	ValidHost        []string
	InvalidHost      []string
}

func newEptTestConfig() *eptTestConfig {
	return &eptTestConfig{}
}

type testEnv struct {
	testDir  string
	testConf string
	testCert string
}

type testData struct {
	contentFile string
	contentCert string
	configExist bool
	certExist   bool
}

func newTestEnv(td *testData, t *testing.T) *testEnv {
	dir, err := ioutil.TempDir("", tempPrefix)
	if err != nil {
		t.Fatal(err)
	}
	var certFile string
	var confFile string

	if td.configExist {
		confFile = filepath.Join(dir, confFileName)
		if err := ioutil.WriteFile(confFile,
			[]byte(td.contentFile), filePerm); err != nil {
			t.Fatal(err)
		}
	}

	if td.certExist {
		certFile = filepath.Join(dir, certFileName)
		if err := ioutil.WriteFile(certFile,
			[]byte(td.contentCert), filePerm); err != nil {
			t.Fatal(err)
		}
	}

	return &testEnv{
		testDir:  dir,
		testConf: confFile,
		testCert: certFile,
	}
}

func (env *testEnv) clean() {
	os.RemoveAll(env.testDir)
}

func validParams(in []string, out map[string]string) bool {
	for _, key := range in {
		delete(out, key)
	}

	if out[caPathName] == "" {
		return false
	}

	delete(out, caPathName)

	if out[caData] == "" {
		return false
	}

	delete(out, caData)

	if len(out) != 0 {
		return false
	}
	return true
}

func TestParsingValidConfig(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		validCert,
		true,
		true}, t)
	defer env.clean()

	out, err := testEMT.ParseConfig(env.testConf)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !validParams(testEMT.keys, out) {
		t.Fatal(errPars)
	}
}

func TestParsingInvalidConfig(t *testing.T) {
	env := newTestEnv(&testData{
		invalidConf,
		validCert,
		true,
		true}, t)
	defer env.clean()

	_, err := testEMT.ParseConfig(env.testConf)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCannotReadCertificateFile(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		"",
		true,
		false}, t)
	defer env.clean()

	_, err := testEMT.ParseConfig(env.testConf)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCertificateIsEmpty(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		"",
		true,
		true}, t)
	defer env.clean()

	_, err := testEMT.ParseConfig(env.testConf)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInvalidCertificate(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		invalidCert,
		true,
		true}, t)
	defer env.clean()

	_, err := testEMT.ParseConfig(env.testConf)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInputFormat(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		validCert,
		true,
		true}, t)
	defer env.clean()

	addParams, err := testEMT.ParseConfig(env.testConf)
	if err != nil {
		t.Fatal(errPars)
	}

	for _, hash := range conf.EptTest.ValidHash {
		_, err := testEMT.Message(
			hash,
			conf.EptTest.ValidHost[0],
			conf.EptTest.ValidHost[1],
			"",
			"",
			addParams,
		)
		if err != nil {
			t.Fatal(errGen)
		}
	}

	for _, hash := range conf.EptTest.InvalidHash {
		_, err := testEMT.Message(
			hash,
			conf.EptTest.ValidHost[0],
			conf.EptTest.ValidHost[1],
			"",
			"",
			addParams,
		)
		if err == nil {
			t.Fatal(errGen)
		}
	}

	for _, host := range conf.EptTest.ValidHost {
		_, err := testEMT.Message(
			conf.EptTest.ValidHash[0],
			host,
			conf.EptTest.ValidHost[0],
			"",
			"",
			addParams,
		)
		if err != nil {
			t.Fatal(errGen)
		}
	}

	for _, host := range conf.EptTest.InvalidHost {
		_, err := testEMT.Message(
			conf.EptTest.ValidHash[0],
			host,
			conf.EptTest.ValidHost[0],
			"",
			"",
			addParams,
		)
		if err == nil {
			t.Fatal(errGen)
		}
	}
}

func TestGenerateCorrectMessage(t *testing.T) {
	env := newTestEnv(&testData{
		validConf,
		validCert,
		true,
		true}, t)
	defer env.clean()

	addParams, err := testEMT.ParseConfig(env.testConf)
	if err != nil {
		t.Fatal(errPars)
	}

	pattern := &EndpointMessage{
		TemplateHash:           conf.EptTest.ValidHash[0],
		PaymentReceiverAddress: conf.EptTest.ValidHost[0],
		ServiceEndpointAddress: conf.EptTest.ValidHost[1],
		AdditionalParams:       addParams,
	}

	candidate := new(EndpointMessage)

	msg, err := testEMT.Message(
		pattern.TemplateHash,
		pattern.PaymentReceiverAddress,
		pattern.ServiceEndpointAddress,
		"",
		"",
		addParams,
	)
	if err != nil {
		t.Fatal(errGen)
	}

	if err := json.Unmarshal(msg, &candidate); err != nil {
		t.Fatal(errGen)
	}

	if !reflect.DeepEqual(pattern, candidate) {
		t.Fatal(errGen)
	}
}

func TestMain(m *testing.M) {
	conf.EptTest = newEptTestConfig()
	util.ReadTestConfig(&conf)
	testEMT = NewEndpointMessageTemplate(conf.EptTest.ExportConfigKeys)

	os.Exit(m.Run())
}
