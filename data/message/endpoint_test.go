package message

import (
	"github.com/privatix/dappctrl/data/message/templates"
	"github.com/privatix/dappctrl/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	env  *testEnv
	conf = &config{}
)

type config struct{}

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

func TestMain(m *testing.M) {
	util.ReadTestConfig(&conf)
}

func newMessage() *EndpointMessageTemplate {
	msg, err := NewEndpointMessageTemplate(
		"0x1234",
		"receiver.example.com:8888",
		"endpoint.example.com:9999",
		"",
		"")
	if err != nil {
		panic(err)
	}
	return msg
}

func newTestEnv(td *testData, t *testing.T) {
	dir, err := ioutil.TempDir("", "message_test")
	if err != nil {
		t.Fatal(err)
	}
	var certFile string
	var confFile string

	if td.configExist {
		confFile = filepath.Join(dir, "server.conf")
		if err := ioutil.WriteFile(confFile, []byte(td.contentFile), 0666); err != nil {
			t.Fatal(err)
		}
	}

	if td.certExist {
		certFile = filepath.Join(dir, "ca.crt")
		if err := ioutil.WriteFile(certFile, []byte(td.contentCert), 0666); err != nil {
			t.Fatal(err)
		}
	}

	env = &testEnv{
		testDir:  dir,
		testConf: confFile,
		testCert: certFile,
	}
}

func cleanTestEnv(t *testing.T) {
	if env != nil && env.testDir != "" {
		if err := os.RemoveAll(env.testDir); err != nil {
			t.Fatal(err)
		}
	}
	env = nil
}

func TestNewEndpointMessageTemplate(t *testing.T) {
	t.Run("valid input params", func(t *testing.T) {
		msg, err := NewEndpointMessageTemplate(
			"0x1234",
			"receiver.example.com:8888",
			"endpoint.example.com:9999",
			"",
			"")
		if err != nil {
			t.Fatal(err)
		}
		if msg == nil {
			t.Fatal("function did not return message")
		}
	})

	t.Run("hash parameter is nill", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"",
			"receiver.example.com:8888",
			"endpoint.example.com:9999",
			"",
			"")
		if err == nil || err.Error() != ErrInput {
			t.Fatal("error is wrong")
		}
	})

	t.Run("receiver parameter is null", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			"",
			"endpoint.example.com:9999",
			"",
			"")
		if err == nil || err.Error() != ErrInput {
			t.Fatal("error is wrong")
		}
	})

	t.Run("endpoint parameter is null", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			"receiver.example.com:8888",
			"",
			"",
			"")
		if err == nil || err.Error() != ErrInput {
			t.Fatal("error is wrong")
		}
	})

	t.Run("receiver parameter have empty host", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			":8888",
			"endpoint.example.com:9999",
			"",
			"")
		if err == nil || err.Error() != ErrReceiver {
			t.Fatal("error is wrong")
		}
	})

	t.Run("receiver parameter have empty port", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			"receiver.example.com:",
			"endpoint.example.com:9999",
			"",
			"")
		if err == nil || err.Error() != ErrReceiver {
			t.Fatal("error is wrong")
		}
	})

	t.Run("endpoint parameter have empty host", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			"receiver.example.com:8888",
			":9999",
			"",
			"")
		if err == nil || err.Error() != ErrEndpoint {
			t.Fatal("error is wrong")
		}
	})

	t.Run("endpoint parameter have empty host", func(t *testing.T) {
		_, err := NewEndpointMessageTemplate(
			"0x1234",
			"receiver.example.com:8888",
			"endpoint.example.com:",
			"",
			"")
		if err == nil || err.Error() != ErrEndpoint {
			t.Fatal("error is wrong")
		}
	})

}

func TestEndpointMessageTemplate_ParsParamsFromConfig(t *testing.T) {
	t.Run("valid input params", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFullExample,
			contentCert: templates.CertValidExample,
			configExist: true,
			certExist:   true,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		result, err := testMessage.ParsParamsFromConfig(env.testConf)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil {
			t.Fatal("function did not return result")
		}
	})

	t.Run("file is empty", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFullExample,
			contentCert: templates.CertValidExample,
			configExist: true,
			certExist:   true,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		_, err := testMessage.ParsParamsFromConfig("")
		if err == nil || err.Error() != ErrFilePathIsEmpty {
			t.Fatal("error is wrong")
		}

	})

	t.Run("file is wrong", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFullExample,
			contentCert: templates.CertValidExample,
			configExist: true,
			certExist:   true,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		_, err := testMessage.ParsParamsFromConfig("fakeFile")
		if err == nil || !strings.Contains(err.Error(), ErrParsingLines) {
			t.Fatal("error is wrong")
		}
	})

	t.Run("server config does not have a certificate", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFakeCertificate,
			contentCert: templates.CertValidExample,
			configExist: true,
			certExist:   true,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		_, err := testMessage.ParsParamsFromConfig(env.testConf)
		if err == nil || !strings.Contains(err.Error(), ErrCertNotExist) {
			t.Fatal("error is wrong")
		}
	})

	t.Run("cannot read certificate file", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFullExample,
			contentCert: "",
			configExist: true,
			certExist:   false,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		_, err := testMessage.ParsParamsFromConfig(env.testConf)
		if err == nil || !strings.Contains(err.Error(), ErrCertCanNotRead) {
			t.Fatal("error is wrong")
		}
	})

	t.Run("certificates is empty", func(t *testing.T) {
		td := &testData{
			contentFile: templates.OpenVpnConfServerFullExample,
			contentCert: "",
			configExist: true,
			certExist:   true,
		}

		newTestEnv(td, t)
		defer cleanTestEnv(t)

		testMessage := newMessage()
		_, err := testMessage.ParsParamsFromConfig(env.testConf)
		if err == nil || !strings.Contains(err.Error(), ErrCertIsNull) {
			t.Fatal("error is wrong")
		}
	})

}
