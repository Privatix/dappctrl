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
)

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
	Keepalive     string // todo : may be needed in the future [maxim]
}

// New returns config object
func New(srvAddr, srvPort string, additionalParams []byte) (*Config, error) {
	if !isHost(srvAddr) {
		return nil, ErrInput
	}

	if err := util.ValidateFormat(util.FormatNetworkPort, srvPort); err != nil {
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
		case "proto":
			config.Proto = val
		case "cipher":
			config.Cipher = val
		case "ping-restart":
			config.PingRestart = val
		case "ping":
			config.Ping = val
		case "connect-retry":
			config.ConnectRetry = val
		case "caData":
			config.Ca = val
		case "comp-lzo":
			config.CompLZO = "comp-lzo"
		}
	}

	return config, nil

}

// GetText injects config values into custom template
func (c *Config) GetText(tpl string) (string, error) {
	t := template.New("config")

	t, err := t.Parse(tpl)
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
