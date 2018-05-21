package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"text/template"

	"github.com/rakyll/statik/fs"

	// This is necessary for statik.
	_ "github.com/privatix/dappctrl/statik"
	"github.com/privatix/dappctrl/util"
)

const (
	filePerm    = 0644
	autogen     = "autogen"
	autogenTest = " # autogenerate option"
	clientTpl   = "/ovpn/templates/client-config.tpl"
)

const (
	nameProto        = "proto"
	nameCipher       = "cipher"
	namePingRestart  = "ping-restart"
	namePing         = "ping"
	nameConnectRetry = "connect-retry"
	nameCaData       = "caData"
	nameCompLZO      = "comp-lzo"
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

var defaultConfig = &CConf{
	Proto:         defaultProto,
	Cipher:        defaultCipher,
	ServerAddress: defaultServerAddress,
	Port:          defaultServerPort,
	PingRestart:   defaultPingRestart,
	Ping:          defaultPing,
	ConnectRetry:  defaultConnectRetry,
}

// CConf OpenVpn client model config
type CConf struct {
	Proto         string
	Cipher        string
	ServerAddress string
	Port          string
	PingRestart   string
	Ping          string
	ConnectRetry  string
	Ca            string
	CompLZO       string
}

// ClientConfig returns config object
func ClientConfig(srvAddr, srvPort string,
	additionalParams []byte) (*CConf, error) {
	if !isHost(srvAddr) || !util.IsNetPort(srvPort) {
		return nil, ErrInput
	}

	var params map[string]string

	if err := json.Unmarshal(additionalParams, &params); err != nil {
		return nil, err
	}

	config := defaultConfig

	config.ServerAddress = srvAddr
	config.Port = srvPort

	for key, val := range params {
		switch key {
		case nameProto:
			config.Proto = val
		case nameCipher:
			config.Cipher = val
		case namePingRestart:
			config.PingRestart = val
		case namePing:
			config.Ping = val
		case nameConnectRetry:
			config.ConnectRetry = val
		case nameCaData:
			config.Ca = val
		case nameCompLZO:
			config.CompLZO = nameCompLZO
		}
	}

	return config, nil

}

// Generate injects config values into custom template
func (c *CConf) Generate(tpl string) (string, error) {
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

// SaveToFile reads ClientConfig template
// and writes result to destination file
func (c *CConf) SaveToFile(destPath string) error {
	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	tpl, err := statikFS.Open(clientTpl)
	if err != nil {
		return err
	}
	defer tpl.Close()

	data, err := ioutil.ReadAll(tpl)
	if err != nil {
		return err
	}

	str, err := c.Generate(string(data))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destPath, []byte(str), filePerm)
}

func autogenFu() string {
	return autogenTest
}
