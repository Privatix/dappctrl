package msg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const (
	defaultAccessFile     = "access.ovpn"
	defaultCipher         = "AES-256-CBC"
	defaultConnectRetry   = "5"
	defaultPing           = "10"
	defaultPingRestart    = "25"
	defaultProto          = "tcp-client"
	defaultServerAddress  = "127.0.0.1"
	defaultServerPort     = "443"
	defaultManagementPort = 7506

	compLZOName = "comp-lzo"
	protoName   = "proto"

	clientConfigName     = "client.ovpn"
	clientConfigTemplate = "/ovpn/templates/client-config.tpl"
	clientTemplateName   = "clientVpnConfig"

	tcp       = "tcp"
	tcpServer = "tcp-server"
)

// specific adapter options
const (
	VpnManagementPort = "vpnManagementPort"
)

var (
	vpnConfigTpl = template.New(clientTemplateName)
)

type vpnClient struct {
	Ca             string `json:"caData"`
	Cipher         string `json:"cipher"`
	ConnectRetry   string `json:"connect-retry"`
	CompLZO        string `json:"comp-lzo"`
	Ping           string `json:"ping"`
	PingRestart    string `json:"ping-restart"`
	Port           string `json:"port"`
	Proto          string `json:"proto"`
	ServerAddress  string `json:"-"`
	AccessFile     string `json:"-"`
	ManagementPort uint16 `json:"-"`
}

type service struct{ logger log.Logger }

func defaultVpnConfig() *vpnClient {
	return &vpnClient{
		Cipher:         defaultCipher,
		ConnectRetry:   defaultConnectRetry,
		Ping:           defaultPing,
		PingRestart:    defaultPingRestart,
		Port:           defaultServerPort,
		Proto:          defaultProto,
		ServerAddress:  defaultServerAddress,
		AccessFile:     defaultAccessFile,
		ManagementPort: defaultManagementPort,
	}
}

func (s *service) fillClientConfig(serviceEndpointAddress string,
	additionalParams []byte) (*vpnClient, error) {
	if !util.IsHostname(serviceEndpointAddress) &&
		!util.IsIPv4(serviceEndpointAddress) {
		s.logger.Add("serviceEndpointAddress",
			serviceEndpointAddress).Error(
			ErrServiceEndpointAddr.Error())
		return nil, ErrServiceEndpointAddr
	}

	cfg := defaultVpnConfig()

	if err := json.Unmarshal(additionalParams, cfg); err != nil {
		s.logger.Error(err.Error())
		return nil, ErrDecodeParams
	}

	cfg.ServerAddress = serviceEndpointAddress
	cfg.Proto = proto(additionalParams)

	if existParam(compLZOName, additionalParams) {
		cfg.CompLZO = compLZOName
	}

	return cfg, nil
}

func (s *service) genClientConfig(text string,
	data interface{}) ([]byte, error) {
	tpl, err := vpnConfigTpl.Parse(text)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, ErrParseConfigTemplate
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		s.logger.Error(err.Error())
		return nil, ErrCombineConfigTemplate
	}

	return buf.Bytes(), nil
}

func configDestination(dir string) string {
	return filepath.Join(dir, clientConfigName)
}

func accessDestination(dir string) string {
	return filepath.Join(dir, defaultAccessFile)
}

func (s *service) makeClientConfig(dir, serviceEndpointAddress string,
	params []byte, options map[string]interface{}) error {
	// fill client configuration from service endpoint address and
	// and parameters received from a agent
	cfg, err := s.fillClientConfig(serviceEndpointAddress, params)
	if err != nil {
		return err
	}

	// add full path to a access file to the configuration
	cfg.AccessFile = accessDestination(dir)

	// add vpn management port to the configuration
	mPort, ok := options[VpnManagementPort]
	if ok {
		if port, ok := mPort.(uint16); ok {
			cfg.ManagementPort = port
		}
	}

	data, err := readFileFromVirtualFS(clientConfigTemplate)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	// fill configuration template
	config, err := s.genClientConfig(string(data), cfg)
	if err != nil {
		return err
	}

	err = writeFile(configDestination(dir), config)
	if err != nil {
		s.logger.Error(err.Error())
		return ErrCreateConfig
	}
	return nil
}

// makes access file with username and password
func makeAccess(dir, username, password string) error {
	data := fmt.Sprintf("%s\n%s\n", username, password)
	return writeFile(accessDestination(dir), []byte(data))
}

// MakeFiles creates configuration files for the product.
func MakeFiles(logger log.Logger, dir, serviceEndpointAddress, username,
	password string, params []byte, options map[string]interface{}) error {
	s := &service{}

	configDst := configDestination(dir)
	accessDst := accessDestination(dir)

	// if the target directory does not exists,
	// then creates target directory
	if notExist(dir) {
		if err := makeDir(dir); err != nil {
			logger.Add("directory", dir).Error(err.Error())
			return ErrCreateDir
		}
	} else {
		// if the configuration file and the access file exist,
		// then complete the function execution
		if checkFile(configDst) && checkFile(accessDst) {
			return nil
		}
	}

	// if the configuration file does not exists,
	// then makes and fill client configuration file
	if !checkFile(configDst) {
		if err := s.makeClientConfig(dir, serviceEndpointAddress,
			params, options); err != nil {
			return err
		}
	}

	// if the access file does not exists,
	// then makes and fill access file
	if !checkFile(accessDst) {
		if err := makeAccess(dir, username, password); err != nil {
			s.logger.Add("directory",
				dir).Error(err.Error())
			return ErrCreateAccessFile
		}
	}
	return nil
}

func variables(data []byte) (v map[string]string) {
	v = make(map[string]string)
	json.Unmarshal(data, &v)
	return
}

func existParam(key string, data []byte) bool {
	v := variables(data)

	if _, ok := v[key]; !ok {
		return false
	}
	return true
}

func proto(data []byte) string {
	v := variables(data)

	val, ok := v[protoName]
	if !ok {
		return defaultProto
	}

	// if proto = tcp-server or tcp then replaces to tcp-client
	if val == tcpServer || val == tcp {
		return defaultProto
	}
	return val
}

// SpecificOptions returns specific options for dappvpn.
// These options will be used to create a product configuration.
func SpecificOptions(monConfig *mon.Config) (
	options map[string]interface{}) {
	options = make(map[string]interface{})

	// TODO(maxim) now we only have VpnManagementPort parameter for `VPN client` product
	// reads OpenVpn management interface address from configuration.
	params := strings.Split(monConfig.Addr, ":")
	if len(params) != 2 {
		return options
	}

	// reads OpenVpn management interface server port from configuration.
	port, err := strconv.ParseUint(params[1], 10, 16)
	if err != nil {
		return options
	}

	options[VpnManagementPort] = uint16(port)
	return options
}
