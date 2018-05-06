package ept

import (
	"io/ioutil"
	"strings"

	"github.com/privatix/dappctrl/util"
)

// ValidNetworkAddress If string includes the address and port
// and has a correct format (address:port), then returns true
func ValidNetworkAddress(addr string) bool {
	data := strings.Split(strings.TrimSpace(addr), ":")
	// check data length
	if len(data) != 2 {
		return false
	}

	host := strings.TrimSpace(data[0])
	port := strings.TrimSpace(data[1])

	// check host and address length
	if len(host) == 0 || len(port) == 0 {
		return false
	}

	// check host format
	if !isHost(host) {
		return false
	}

	// check port format
	if util.ValidateFormat(util.FormatNetworkPort, port) != nil {
		return false
	}
	return true
}

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
