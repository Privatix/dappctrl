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

	if !util.IsTLSCert(string(mainCertPEMBlock)) {
		return "", ErrCertNotFound
	}

	return string(mainCertPEMBlock), nil
}

func isHost(host string) bool {
	if util.IsHostname(host) || util.IsIPv4(host) {
		return true
	}
	return false
}
