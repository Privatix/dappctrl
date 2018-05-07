package ept

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"text/template"

	"github.com/privatix/dappctrl/ept/templates/ovpn"
	"github.com/privatix/dappctrl/util"
)

const (
	filePerm = 0644
	autogen  = "autogen"
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

var defaultConfig = &Config{
	Proto:         defaultProto,
	Cipher:        defaultCipher,
	ServerAddress: defaultServerAddress,
	Port:          defaultServerPort,
	PingRestart:   defaultPingRestart,
	Ping:          defaultPing,
	ConnectRetry:  defaultConnectRetry,
}

// Config OpenVpn client model config
type Config struct {
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

// New returns config object
func New(srvAddr, srvPort string, additionalParams []byte) (*Config, error) {
	if !isHost(srvAddr) {
		return nil, ErrInput
	}

	if err := util.ValidateFormat(util.FormatNetworkPort,
		srvPort); err != nil {
		return nil, err
	}

	var params map[string]string

	if err := json.Unmarshal(additionalParams, &params); err != nil {
		return nil, err
	}

	config := new(Config)

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

	//config.normalize()

	return config, nil

}

// GetText injects config values into custom template
func (c *Config) GetText(tpl string) (string, error) {
	t := template.New("config")

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

// SaveToFile reads ClientConfig teamplate
// and writes result to destination file
func (c *Config) SaveToFile(destPath string) error {
	str, err := c.GetText(ovpn.ClientConfig)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destPath, []byte(str), filePerm)
}

func (c *Config) normalize() {
	if c.Proto == "" {
		c.Proto = defaultConfig.Proto
	}

	if c.Cipher == "" {
		c.Cipher = defaultConfig.Cipher
	}

	if c.ServerAddress == "" {
		c.ServerAddress = defaultConfig.ServerAddress
	}

	if c.Port == "" {
		c.Port = defaultConfig.Port
	}

	if c.PingRestart == "" {
		c.PingRestart = defaultConfig.PingRestart
	}

	if c.Ping == "" {
		c.Ping = defaultConfig.Ping
	}

	if c.ConnectRetry == "" {
		c.ConnectRetry = defaultConfig.ConnectRetry
	}
}

func autogenFu() string {
	return " # autogenerate option"
}
