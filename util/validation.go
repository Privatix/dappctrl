package util

import (
	"crypto/tls"
	"encoding/pem"
	"net"
	"strconv"

	"github.com/xeipuuv/gojsonschema"
)

const certificate = "CERTIFICATE"

// IsIPv4 checks if this is a valid IPv4
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil
}

// IsHostname checks if this is a hostname
func IsHostname(s string) bool {
	addrs, err := net.LookupHost(s)
	if err != nil || len(addrs) == 0 {
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

// ValidateJSON validates a given JSON against a given schema.
func ValidateJSON(schema, data []byte) bool {
	sloader := gojsonschema.NewBytesLoader(schema)
	dloader := gojsonschema.NewBytesLoader(data)
	result, err := gojsonschema.Validate(sloader, dloader)
	return err == nil && result.Valid() && len(result.Errors()) == 0
}
