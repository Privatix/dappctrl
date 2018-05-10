package config

import (
	"io/ioutil"

	"github.com/privatix/dappctrl/util"
)

// ParseCertFromFile Parsing TLS certificate from file
func ParseCertFromFile(caCertPath string) (string, error) {
	mainCertPEMBlock, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return "", ErrCertCanNotRead
	}

	if err = util.ValidateFormat(util.FormatTLSCertificate,
		string(mainCertPEMBlock)); err != nil {
		return "", err
	}

	return string(mainCertPEMBlock), nil
}

func isHost(host string) bool {
	if util.ValidateFormat(util.FormatIP, host) == nil ||
		util.ValidateFormat(util.FormatIPv4, host) == nil ||
		util.ValidateFormat(util.FormatIPv6, host) == nil ||
		util.ValidateFormat(util.FormatHostname, host) == nil {
		return true
	}
	return false
}
