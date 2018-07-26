package msg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/privatix/dappctrl/util"
)

const (
	serviceEndpointAddress = "example.com"
	password               = "secret"

	caDataKey = "caData"
	remoteKey = "remote"
	portKey   = "port"
)

var (
	username = util.NewUUID()
)

func readStatikFile(t *testing.T, name string) []byte {
	data, err := readFileFromVirtualFS(name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func copyStrStrMap(params map[string]string) (dst map[string]string) {
	dst = make(map[string]string)

	for k, v := range params {
		dst[k] = v
	}
	return dst
}

func testAdditionalParams(t *testing.T, parameters map[string]string) map[string]string {
	ca := readStatikFile(t, sampleCa)

	result := copyStrStrMap(parameters)
	result[caDataKey] = string(ca)
	return result
}

func checkAccess(t *testing.T, file, username, password string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte(fmt.Sprintf("%s\n%s\n", username, password))

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("expected %s, got %s", expected, data)
	}
}

func checkCA(t *testing.T, config string, ca []byte) {
	configData, err := ioutil.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}

	start := strings.Index(string(configData), `<ca>`)
	stop := strings.LastIndex(string(configData), `</ca>`)

	data := configData[start+5 : stop]

	if !reflect.DeepEqual(ca, data) {
		t.Fatalf("expected %s, got %s", ca, data)
	}
}

func checkConf(t *testing.T, config string, keys []string) {
	keys = append(keys, remoteKey)

	result, err := vpnParams(config, keys)
	if err != nil {
		t.Fatal(err)
	}

	// checks special argument "port"
	val, ok := result[remoteKey]
	if !ok {
		t.Fatal(`special argument "remote" not exists`)
	}

	if !strings.HasSuffix(val, parameters[portKey]) {
		t.Fatal(`special argument "port" not exists`)
	}

	// clears the map of special parameters
	delete(result, remoteKey)

	// adds the just-tested parameter to the resulting map
	result[portKey] = parameters[portKey]

	// checks special parameter "proto"
	if result[protoName] != defaultProto {
		t.Fatal(`special argument "proto" must be "tcp-client"`)
	}

	// Changes the just-tested parameter to the resulting map.
	// On the client "tcp" parameter is replaced by "tcp-client"
	result[protoName] = parameters[protoName]

	// others parameters not change
	if !reflect.DeepEqual(parameters, result) {
		t.Fatal("result parameters not equals initial parameters")
	}
}

func TestMakeFiles(t *testing.T) {
	rootDir, err := ioutil.TempDir("", util.NewUUID())
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(rootDir)

	params := testAdditionalParams(t, parameters)

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}

	accessFile := filepath.Join(rootDir, defaultAccessFile)
	confFile := filepath.Join(rootDir, clientConfigName)

	if err := MakeFiles(rootDir, serviceEndpointAddress, username,
		password, data, SpecificOptions(
			conf.VPNMonitor)); err != nil {
		t.Fatal(err)
	}

	checkAccess(t, accessFile, username, password)
	checkCA(t, confFile, []byte(params[caDataKey]))
	checkConf(t, confFile, parameterKeys(parameters))
}
