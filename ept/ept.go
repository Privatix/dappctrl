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

type message struct {
	templateHash           string
	paymentReceiverAddress string
	serviceEndpointAddress string
	username               string
	password               string
	additionalParams       map[string]string
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

// Message generates new Endpoint message into JSON format
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
		return nil, ErrEndpoint
	}

	if !e.validNetworkAddress(endpoint) {
		return nil, ErrEndpoint
	}

	if util.ValidateFormat(util.FormatEthHash, hash) != nil {
		return nil, ErrHash
	}

	return json.Marshal(&message{
		templateHash:           hash,
		paymentReceiverAddress: receiver,
		serviceEndpointAddress: endpoint,
		username:               username,
		password:               password,
		additionalParams:       additionalParams,
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
	if util.ValidateFormat(util.FormatIP, host) != nil &&
		util.ValidateFormat(util.FormatHostname, data[0]) != nil {
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

	ca := results[caNameFromConfig]
	if ca == "" {
		return nil, ErrCertNotExist
	}

	certPath := filepath.Dir(filePath) + string(os.PathSeparator) + ca

	cert, err := e.ParseCert(certPath)
	if err != nil {
		return nil, err
	}
	results[caData] = cert
	results[caPathName] = certPath
	return results, nil
}

func (e *EndpointMessageTemplate) parseLine(
	keys map[string]bool, line string) (string, string, bool) {
	str := strings.TrimSpace(line)
	for key := range keys {
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
