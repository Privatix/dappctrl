package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/rakyll/statik/fs"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	// This is necessary for statik.
	_ "github.com/privatix/dappctrl/statik"
)

const (
	filePerm = 0644
	pathPerm = 0755

	autogen     = "autogen"
	autogenTest = " # autogenerate option"

	clientTpl        = "/ovpn/templates/client-config.tpl"
	clientConfName   = "client.ovpn"
	clientAccessName = "access.ovpn"
)

const (
	nameCompLZO = "comp-lzo"
)

const (
	defaultProto         = "tcp"
	defaultCipher        = "AES-256-CBC"
	defaultServerAddress = "127.0.0.1"
	defaultServerPort    = "443"
	defaultPingRestart   = "10"
	defaultPing          = "10"
	defaultConnectRetry  = "2 120"
)

var defaultConfig = &cConf{
	Proto:         defaultProto,
	Cipher:        defaultCipher,
	ServerAddress: defaultServerAddress,
	Port:          defaultServerPort,
	PingRestart:   defaultPingRestart,
	Ping:          defaultPing,
	ConnectRetry:  defaultConnectRetry,
}

// cConf OpenVpn client model config
type cConf struct {
	Proto         string `json:"proto"`
	Cipher        string `json:"cipher"`
	ServerAddress string
	Port          string `json:"port"`
	PingRestart   string `json:"ping-restart"`
	Ping          string `json:"ping"`
	ConnectRetry  string `json:"connect-retry"`
	Ca            string `json:"caData"`
	CompLZO       string `json:"comp-lzo"`
}

// ConfDeployer is needed to store information about the directory in which
// the directories with configuration files are stored
type ConfDeployer struct {
	rootDir string
}

// NewConfDeployer creates a new ConfDeployer object
func NewConfDeployer(rootDir string) *ConfDeployer {
	return &ConfDeployer{rootDir}
}

// Deploy creates target directory, the name is equivalent to channel.ID or
// endpoint.Channel. In target directory, two files are created ("client.ovpn",
//"access.ovpn"): 1) "client.ovpn" - the file is filled with
// parameters from "srvAddr" (server host or ip address) and params (OpenVpn
// server configuration parameters) 2) "access.ovpn" - the file is filled
// "login" and "pass" parameters
func (d *ConfDeployer) Deploy(record reform.Record, srvAddr string,
	login, pass string, params []byte) (string, error) {
	var target string

	if record == nil || !isHost(srvAddr) || login == "" ||
		pass == "" || len(params) == 0 {
		return "", ErrInput
	}

	switch r := record.(type) {
	case *data.Channel:
		if r.ID == "" {
			return "", ErrInput
		}

		target = filepath.Join(d.rootDir, r.ID)
		if err := createPath(target); err != nil {
			return "", err
		}
	case *data.Endpoint:
		if r.Channel == "" {
			return "", ErrInput
		}

		target = filepath.Join(d.rootDir, r.Channel)
		if err := createPath(target); err != nil {
			return "", err
		}
	default:
		return "", ErrInput
	}

	if err := d.deploy(target, srvAddr, login, pass, params); err != nil {
		return "", err
	}

	return target, nil
}

func (d *ConfDeployer) deploy(targetDir, srvAddr string,
	login, pass string, params []byte) error {
	cfg, err := clientConfig(srvAddr, params)
	if err != nil {
		return err
	}

	if isNotExist(targetDir) {
		if err := createPath(targetDir); err != nil {
			return err
		}
	}

	dir := filepath.Join(targetDir)

	if err := cfg.saveToFile(filepath.Join(dir,
		clientConfName)); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(dir, clientAccessName),
		[]byte(fmt.Sprintf("%s\n%s\n", login, pass)),
		filePerm)
}

func autogenFu() string {
	return autogenTest
}

func checkParam(key string, data []byte) bool {
	var vals map[string]string

	if err := json.Unmarshal(data, &vals); err != nil {
		return false
	}

	if _, ok := vals[key]; !ok {
		return false
	}
	return true
}

// generate injects config values into custom template
func (c *cConf) generate(tpl string) (string, error) {
	t := template.New(clientTpl)

	t, err := t.Funcs(
		template.FuncMap{autogen: autogenFu}).Parse(tpl)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	if err := t.Execute(buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// saveToFile reads clientConfig template
// and writes result to destination file
func (c *cConf) saveToFile(destPath string) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	tpl, err := statikFS.Open(clientTpl)
	if err != nil {
		return err
	}
	defer tpl.Close()

	d, err := ioutil.ReadAll(tpl)
	if err != nil {
		return err
	}

	str, err := c.generate(string(d))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destPath, []byte(str), filePerm)
}

// clientConfig returns config object
func clientConfig(srvAddr string,
	additionalParams []byte) (*cConf, error) {
	if !isHost(srvAddr) {
		return nil, ErrInput
	}

	config := defaultConfig

	if err := json.Unmarshal(additionalParams, config); err != nil {
		return nil, err
	}

	config.ServerAddress = srvAddr

	if checkParam(nameCompLZO, additionalParams) {
		config.CompLZO = nameCompLZO
	}

	return config, nil

}
