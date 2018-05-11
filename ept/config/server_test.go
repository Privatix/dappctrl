package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/privatix/dappctrl/util"
)

var (
	conf struct {
		EptTest *eptTestConfig
	}
)

const (
	errPars     = "incorrect parsing test"
	samplesPath = "samples"
)

type eptTestConfig struct {
	Template            string
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
	out, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfValidCaValid),
		true, conf.EptTest.ExportConfigKeys)
	if err != nil {
		t.Fatal(err.Error())
	}

	r, _ := json.Marshal(out)

	t.Log(string(r))

	if !validParams(conf.EptTest.ExportConfigKeys, out) {
		t.Fatal(errPars)
	}
}

func TestParsingInvalidConfig(t *testing.T) {
	_, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfInvalid),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCannotReadCertificateFile(t *testing.T) {
	_, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfValidCaNotExist),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestCertificateIsEmpty(t *testing.T) {
	_, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfValidCaEmpty),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestInvalidCertificate(t *testing.T) {
	_, err := ServerConfig(filepath.Join(samplesPath,
		conf.EptTest.ConfValidCaInvalid),
		true, conf.EptTest.ExportConfigKeys)
	if err == nil {
		t.Fatal(errPars)
	}
}

func TestMain(m *testing.M) {
	conf.EptTest = newEptTestConfig()

	util.ReadTestConfig(&conf)

	os.Exit(m.Run())
}
