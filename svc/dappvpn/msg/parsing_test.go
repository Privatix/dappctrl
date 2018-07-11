package msg

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/privatix/dappctrl/util"
)

const (
	configName = "server.ovpn"
)

func parameterKeys(params map[string]string) (keys []string) {
	for k := range params {
		keys = append(keys, k)
	}
	return keys
}

func createTestFile(t *testing.T, file string, params map[string]string) {
	var data []byte
	buf := bytes.NewBuffer(data)

	for k, v := range params {
		if _, err := buf.WriteString(k + " " + v + "\n"); err != nil {
			t.Fatal(err)
		}
	}

	if err := ioutil.WriteFile(file,
		buf.Bytes(), os.ModePerm); err != nil {
		t.Fatal(err)
	}
}

func TestParsingVpnConfigFile(t *testing.T) {
	dir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	config := filepath.Join(dir, configName)

	defer os.RemoveAll(dir)

	createTestFile(t, filepath.Join(dir, configName), parameters)

	result, err := vpnParams(config, parameterKeys(parameters))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(parameters, result) {
		t.Fatal("result parameters not equals initial parameters")
	}
}
