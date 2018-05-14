package util

import (
	"crypto/tls"
	"encoding/pem"
	"net"
	"regexp"
	"strconv"
)

const certificate = "CERTIFICATE"

var (
	// Regular expression used to validate RFC1035 hostnames*/
	hostnameRegex = regexp.MustCompile(
		`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

	hostnameRegex2 = regexp.MustCompile(
		`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)

	// Simple regular expression for IPv4 values,
	// more rigorous checking is done via net.ParseIP
	ipv4Regex = regexp.MustCompile(
		`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
)

// IsIPv4 checks if this is a valid IPv4
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	if !ipv4Regex.MatchString(s) {
		return false
	}
	return true
}

// IsHostname checks if this is a hostname
func IsHostname(s string) bool {
	if !hostnameRegex.MatchString(s) &&
		!hostnameRegex2.MatchString(s) {
		return false
	}
	return true
}

// IsNetPort checks if this is a valid net port
func IsNetPort(str string) bool {
	if _, err := strconv.ParseUint(
		str, 10, 16); err != nil {
		return false
	}
	return true
}

// IsTLSCert if block is one or more
// TLS certificates then function returns true
func IsTLSCert(block string) bool {
	var cert tls.Certificate

	pemBlock := []byte(block)

	for {
		var derBlock *pem.Block
		derBlock, pemBlock = pem.Decode(pemBlock)
		if derBlock == nil {
			break
		}

		if derBlock.Type == certificate {
			cert.Certificate =
				append(cert.Certificate, derBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		return false
	}

	return true
}
