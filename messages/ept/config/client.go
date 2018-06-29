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
	autogen          = "autogen"
	autogenTest      = " # autogenerate option"
	clientAccessName = "access.ovpn"
	clientConfName   = "client.ovpn"
	clientTpl        = "/ovpn/templates/client-config.tpl"
	filePerm         = 0644
	pathPerm         = 0755

	nameCompLZO = "comp-lzo"
	nameProto   = "proto"

	defaultCipher       = "AES-256-CBC"
	defaultConnectRetry = "2 120"
	defaultPing         = "10"
	defaultPingRestart  = "10"
	defaultProto        = "tcp-client"
	defaultAccessFile   = "access.ovpn"

	// DefaultServerAddress default OpenVpn server address.
	DefaultServerAddress = "127.0.0.1"
	defaultServerPort    = "443"
)

var defaultConfig = newVpnConfig()

// Config is a configuration for VPN client.
type Config struct {
	ConfigDir string
}

// vpnConf OpenVpn client model config
type vpnConf struct {
	Ca            string `json:"caData"`
	Cipher        string `json:"cipher"`
	ConnectRetry  string `json:"connect-retry"`
	CompLZO       string `json:"comp-lzo"`
	Ping          string `json:"ping"`
	PingRestart   string `json:"ping-restart"`
	Port          string `json:"port"`
	Proto         string `json:"proto"`
	ServerAddress string `json:"serverAddress"`
	AccessFile    string
}

func newVpnConfig() *vpnConf {
	return &vpnConf{
		Cipher:        defaultCipher,
		ConnectRetry:  defaultConnectRetry,
		Ping:          defaultPing,
		PingRestart:   defaultPingRestart,
		Port:          defaultServerPort,
		Proto:         defaultProto,
		ServerAddress: DefaultServerAddress,
		AccessFile:    defaultAccessFile,
	}
}

// NewConfig creates a default VPN client configuration.
func NewConfig() *Config {
	return &Config{"/etc/openvpn/config"}
}

// DeployConfig creates target directory, the name is equivalent
// to endpoint.Channel. In target directory, two files are created
// ("client.ovpn", "access.ovpn"): 1) "client.ovpn" - the file is filled with
// parameters from endpoint.ServiceEndpointAddress (server host or ip address)
// and endpoint.AdditionalParams (OpenVpn
// server configuration parameters) 2) "access.ovpn" - the file is filled
// endpoint.Channel and endpoint.Password parameters.
// dir - information about the location of the directory in which
// the directories with configuration files are stored.
func DeployConfig(db *reform.DB, endpoint, dir string) error {
	e := new(data.Endpoint)

	if err := db.FindByPrimaryKeyTo(e, endpoint); err != nil {
		return ErrEndpoint
	}

	target := filepath.Join(dir, e.Channel)

	if err := createPath(target); err != nil {
		return err
	}

	save := func(str *string) string {
		if str != nil {
			return *str
		}
		return ""
	}

	return deploy(target, save(e.ServiceEndpointAddress),
		save(e.Username), save(e.Password), e.AdditionalParams)
}

func deploy(targetDir, srvAddr string,
	login, pass string, params []byte) error {
	cfg, err := clientConfig(srvAddr, params)
	if err != nil {
		return err
	}

	if notExist(targetDir) {
		if err := createPath(targetDir); err != nil {
			return err
		}
	}

	dir := filepath.Join(targetDir)

	// set absolute path for  access file
	cfg.AccessFile = filepath.Join(dir, defaultAccessFile)

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
	v := variables(data)

	if _, ok := v[key]; !ok {
		return false
	}
	return true
}

func proto(data []byte) string {
	v := variables(data)

	val, ok := v[nameProto]
	if !ok {
		return defaultProto
	}

	if val == "tcp-server" || val == "tcp" {
		return defaultProto
	}
	return val

}

func variables(data []byte) (v map[string]string) {
	v = make(map[string]string)
	json.Unmarshal(data, &v)
	return
}

// generate injects config values into custom template
func (c *vpnConf) generate(tpl string) (string, error) {
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
func (c *vpnConf) saveToFile(destPath string) error {
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
	additionalParams []byte) (*vpnConf, error) {
	if !isHost(srvAddr) {
		return nil, ErrInput
	}

	config := defaultConfig

	if err := json.Unmarshal(additionalParams, config); err != nil {
		return nil, err
	}

	// if the configuration does not have a server address,
	// then take it from srvAddr
	if (config.ServerAddress == DefaultServerAddress ||
		config.ServerAddress == "") && srvAddr != "" {
		config.ServerAddress = srvAddr
	}

	if checkParam(nameCompLZO, additionalParams) {
		config.CompLZO = nameCompLZO
	}

	config.Proto = proto(additionalParams)

	return config, nil
}
