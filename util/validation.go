package util

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"net"
	"regexp"
	"strconv"
)

// Format defines a validation format.
type Format string

const (
	// FormatHostname defines RFC1035 Internet host names.
	FormatHostname = "hostname"

	// FormatIPv4 defines RFC2373 IPv4 address values.
	FormatIPv4 = "ipv4"

	// FormatIPv6 defines RFC2373 IPv6 address values.
	FormatIPv6 = "ipv6"

	// FormatIP defines RFC2373 IPv4 or IPv6 address values.
	FormatIP = "ip"

	// FormatEthHash defines Ethereum hash values
	FormatEthHash = "ethHash"

	// FormatEthHash defines Ethereum hash values
	FormatEthAddress = "ethAddr"

	// FormatPort defines network port values
	FormatNetworkPort = "netPort"
)

var (
	unknownFormat   = "unknown format %#v"
	invalidValue    = "invalid %s value, %s"
	invalidIPv6     = "\"%s\" is an invalid ipv6 value"
	invalidIPv4     = "\"%s\" is an invalid ipv4 value"
	invalidIP       = "\"%s\" is an invalid %s value"
	invalidHostname = "hostname value '%s' does not match %s"
	invalidEthAddr  = "\"%s\" is an invalid Ethereum address value"
	invalidEthHash  = "\"%s\" is an invalid Ethereum hash value"
	invalidNetPort  = "\"%s\" is an invalid network port value"
)

var (
	// Regular expression used to validate RFC1035 hostnames*/
	hostnameRegex = regexp.MustCompile(
		`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

	// Simple regular expression for IPv4 values,
	// more rigorous checking is done via net.ParseIP
	ipv4Regex = regexp.MustCompile(
		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
)

// ValidateFormat validates a string against a standard format.
// Supported formats are:
//     - "hostname": RFC1035 Internet host name
//     - "ipv4", "ipv6", "ip": RFC2673 and RFC2373 IP address values
//     - "ethHash" - Ethereum hash values
//     - "ethAddr" - Ethereum address values
//     - "netPort" - network port values
func ValidateFormat(f Format, val string) error {
	var err error
	switch f {
	case FormatHostname:
		if !hostnameRegex.MatchString(val) {
			err = fmt.Errorf(invalidHostname,
				val, hostnameRegex.String())
		}
	case FormatIPv4, FormatIPv6, FormatIP:
		ip := net.ParseIP(val)
		if ip == nil {
			err = fmt.Errorf(invalidIP, val, f)
		}
		if f == FormatIPv4 {
			if !ipv4Regex.MatchString(val) {
				err = fmt.Errorf(invalidIPv4, val)
			}
		}
		if f == FormatIPv6 {
			if ipv4Regex.MatchString(val) {
				err = fmt.Errorf(invalidIPv6, val)
			}
		}
	case FormatEthAddress:
		if !common.IsHexAddress(val) {
			err = fmt.Errorf(invalidEthAddr, val)
		}
	case FormatEthHash:
		if !isEthHash(val) {
			err = fmt.Errorf(invalidEthHash, val)
		}
	case FormatNetworkPort:
		if !isPort(val) {
			err = fmt.Errorf(invalidNetPort, val)
		}
	default:
		return fmt.Errorf(unknownFormat, f)
	}
	if err != nil {
		return fmt.Errorf(invalidValue, f, err)
	}
	return nil
}

func isEthHash(s string) bool {
	if hasHexPrefix(s) {
		s = s[2:]
	}
	return len(s) == 2*common.HashLength && isHex(s)
}

func hasHexPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' &&
		(str[1] == 'x' || str[1] == 'X')
}

func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') ||
		('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}

func isHex(str string) bool {
	if len(str)%2 != 0 {
		return false
	}
	for _, c := range []byte(str) {
		if !isHexCharacter(c) {
			return false
		}
	}
	return true
}

func isPort(str string) bool {
	port, err := strconv.Atoi(str)
	if err != nil {
		return false
	}
	if port < 0 || port > 65535 {
		return false
	}
	return true
}
