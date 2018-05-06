package ept

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dappctrl/util"
)

const (
	caNameFromConfig = "ca"
	caPathName       = "caPathName"
	caData           = "caData"
)

// EndpointMessageTemplate structure for Endpoint message template
type EndpointMessageTemplate struct {
	keys []string
}

// EndpointMessage structure
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

// Message generates new Endpoint message into JSON format
func (e *EndpointMessageTemplate) Message(hash, receiver, endpoint, username,
	password string, additionalParams map[string]string) ([]byte, error) {
	if hash == "" || receiver == "" || endpoint == "" {
		return nil, ErrInput
	}

	if !ValidNetworkAddress(receiver) {
		return nil, ErrReceiver
	}

	if !ValidNetworkAddress(endpoint) {
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
// certificate from file.
func (e *EndpointMessageTemplate) ParseConfig(
	filePath string) (map[string]string, error) {
	if filePath == "" {
		return nil, ErrFilePathIsEmpty
	}
	return e.parseConfig(filePath, e.keys)
}

func (e *EndpointMessageTemplate) parseConfig(filePath string,
	keys []string) (map[string]string, error) {
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
			cert, err := ParseCertFromFile(filePath)
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

func (e *EndpointMessageTemplate) parseLine(keys map[string]bool,
	line string) (string, string, bool) {
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
