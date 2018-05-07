package ept

import (
	"encoding/json"
	"os"
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
	errPars     = "incorrect parsing test"
	errGen      = "incorrect generate message test"
	samplesPath = "samples" + string(os.PathSeparator)
)

type eptTestConfig struct {
	ExportConfigKeys    []string
	ValidHash           []string
	InvalidHash         []string
	ValidHost           []string
	InvalidHost         []string
	ConfValidCaValid    string
	ConfInvalid         string
	ConfValidCaInvalid  string
	ConfValidCaEmpty    string
	ConfValidCaNotExist string
}

func newEptTestConfig() *eptTestConfig {
	return &eptTestConfig{}
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
	out, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaValid, true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !validParams(testEMT.keys, out) {
		t.Fatal(errPars)
	}
}

func TestParsingInvalidConfig(t *testing.T) {
	_, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfInvalid, true)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCannotReadCertificateFile(t *testing.T) {
	_, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaNotExist, true)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCertificateIsEmpty(t *testing.T) {
	_, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaEmpty, true)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInvalidCertificate(t *testing.T) {
	_, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaInvalid, true)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInputFormat(t *testing.T) {
	addParams, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaValid, true)
	if err != nil {
		t.Fatal(errPars)
	}

	for _, hash := range conf.EptTest.ValidHash {
		if _, err := testEMT.Message(hash, conf.EptTest.ValidHost[0],
			conf.EptTest.ValidHost[1], "", "",
			addParams); err != nil {
			t.Fatal(errGen)
		}
	}

	for _, hash := range conf.EptTest.InvalidHash {
		if _, err := testEMT.Message(hash, conf.EptTest.ValidHost[0],
			conf.EptTest.ValidHost[1], "", "",
			addParams); err == nil {
			t.Fatal(errGen)
		}
	}

	for _, host := range conf.EptTest.ValidHost {
		if _, err := testEMT.Message(conf.EptTest.ValidHash[0], host,
			conf.EptTest.ValidHost[0], "", "",
			addParams); err != nil {
			t.Fatal(errGen)
		}
	}

	for _, host := range conf.EptTest.InvalidHost {
		if _, err := testEMT.Message(conf.EptTest.ValidHash[0], host,
			conf.EptTest.ValidHost[0], "", "",
			addParams); err == nil {
			t.Fatal(errGen)
		}
	}
}

func TestGenerateCorrectMessage(t *testing.T) {
	addParams, err := testEMT.ParseConfig(samplesPath+
		conf.EptTest.ConfValidCaValid, true)
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

	msg, err := testEMT.Message(pattern.TemplateHash,
		pattern.PaymentReceiverAddress, pattern.ServiceEndpointAddress,
		"", "", addParams)
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
