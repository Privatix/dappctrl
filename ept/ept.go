package ept

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dappctrl/util"
)

const (
	certificate      = "CERTIFICATE"
	caNameFromConfig = "ca"
	caPathName       = "caCertPath"
	caData           = "caData"
)

// EndpointMessageTemplate structure for Endpoint message template
type EndpointMessageTemplate struct {
	keys []string
}

// Endpoint Message structure
//
// TemplateHash - Hash of template that was used to fill this message
//
// PaymentReceiverAddress - Address ("hostname:port")
// of payment receiver. Can be dns or IP.
//
// ServiceEndpointAddress - Address ("hostname:port")
// of service endpoint. Can be dns or IP.
//
// Username - Optional fields for username.
//
// Password - Optional fields for password.
//
// AdditionalParams - all additional parameters stored as inner JSON
type EndpointMessage struct {
	TemplateHash           string
	PaymentReceiverAddress string
	ServiceEndpointAddress string
	Username               string
	Password               string
	AdditionalParams       map[string]string
}

// NewEndpointMessageTemplate creates
// the EndpointMessageTemplate object
// username and password optional fields
func NewEndpointMessageTemplate(
	keys []string) *EndpointMessageTemplate {
	return &EndpointMessageTemplate{
		keys: keys,
	}
}

// EndpointMessage generates new Endpoint message into JSON format
func (e *EndpointMessageTemplate) Message(
	hash,
	receiver,
	endpoint,
	username,
	password string,
	additionalParams map[string]string,
) ([]byte, error) {
	if hash == "" || receiver == "" || endpoint == "" {
		return nil, ErrInput
	}

	if !e.validNetworkAddress(receiver) {
		return nil, ErrReceiver
	}

	if !e.validNetworkAddress(endpoint) {
		return nil, ErrEndpoint
	}

	if util.ValidateFormat(util.FormatEthHash, hash) != nil {
		return nil, ErrHash
	}

	return json.Marshal(&EndpointMessage{
		TemplateHash:           hash,
		PaymentReceiverAddress: receiver,
		ServiceEndpointAddress: endpoint,
		Username:               username,
		Password:               password,
		AdditionalParams:       additionalParams,
	})
}

// ParseConfig parsing OpenVpn config file and parsing
// CA certificate from file.
func (e *EndpointMessageTemplate) ParseConfig(
	filePath string) (map[string]string, error) {
	if filePath == "" {
		return nil, ErrFilePathIsEmpty
	}
	return e.parseConfig(filePath, e.keys)
}

// ParseCert CA certificate from file
func (e *EndpointMessageTemplate) ParseCert(
	caCertPath string) (string, error) {
	mainCertPEMBlock, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return "", ErrCertCanNotRead
	}

	var cert tls.Certificate

	certPEMBlock := mainCertPEMBlock

	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == certificate {
			cert.Certificate =
				append(cert.Certificate, certDERBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		return "", ErrCertIsNull
	}
	return string(mainCertPEMBlock), nil
}

func (e *EndpointMessageTemplate) validNetworkAddress(addr string) bool {

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
	checkHost := func(host string) bool {
		if util.ValidateFormat(util.FormatIP, host) == nil ||
			util.ValidateFormat(util.FormatIPv4, host) == nil ||
			util.ValidateFormat(util.FormatIPv6, host) == nil ||
			util.ValidateFormat(util.FormatHostname, host) == nil {
			return true
		}
		return false
	}
	if !checkHost(host) {
		return false
	}

	// check port format
	if util.ValidateFormat(util.FormatNetworkPort, port) != nil {
		return false
	}
	return true
}

func (e *EndpointMessageTemplate) parseConfig(
	filePath string, keys []string) (map[string]string, error) {
	// check input
	if keys == nil || filePath == "" {
		return nil, ErrInput
	}

	// open config file
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	// delete duplicates
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[strings.TrimSpace(key)] = true
	}

	results := make(map[string]string)

	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		if key, value, add :=
			e.parseLine(keyMap, scanner.Text()); add {
			if key != "" {
				results[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// check ca key
	ca := results[caNameFromConfig]
	if ca == "" {
		return nil, ErrCertNotExist
	}

	// A certificate can be located on an absolute and relative path.
	// If the certificate is found in the file,
	// will return the body of the certificate,
	// the absolute path to the certificate, and true
	pCert := func(paths []string) (string, string, bool) {
		for _, filePath := range paths {
			cert, err := e.ParseCert(filePath)
			if err == nil {
				return cert, filePath, true
			}
		}

		return "", "", false
	}

	// absolute path
	absPath := filepath.Dir(filePath) + string(os.PathSeparator) + ca

	cert, certPath, found := pCert([]string{ca, absPath})
	if !found {
		return nil, ErrCertNotFound
	}

	results[caData] = cert
	results[caPathName] = certPath

	return results, nil
}

func (e *EndpointMessageTemplate) parseLine(
	keys map[string]bool, line string) (string, string, bool) {
	str := strings.TrimSpace(line)

	for key := range keys {
		if !strings.HasPrefix(str, key) {
			continue
		}

		index := strings.Index(str, "#")

		if index == -1 {
			words := strings.Split(str, " ")
			if len(words) == 1 {
				return key, "", true
			}
			value := strings.Join(words[1:], " ")
			return key, value, true
		}

		subStr := strings.TrimSpace(str[:index])

		words := strings.Split(subStr, " ")

		if len(words) == 1 {
			return key, "", true
		}

		value := strings.Join(words[1:], " ")

		return key, value, true
	}
	return "", "", false
}
