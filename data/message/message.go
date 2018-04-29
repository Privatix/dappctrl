package message

import (
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"github.com/pkg/errors"
	"github.com/privatix/dappctrl/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ====== start config =====
// proto tcp # Use protocol tcp for communicating with remote host
// cipher AES‌-256-CBC # Encrypt packets with AES‌-256-CBC algorithm
// ping-restart 10 # Trigger a SIGUSR1 restart after n seconds pass without reception of a ping or other packet from remote.
// ping 10 # Ping remote over the TCP/UDP control channel if no packets have been sent for at least n seconds
// connect-retry 2 120 # take n as the number of seconds to wait between connection retries
// <ca>Server Certification Authority (CA) certificate goes here</ca> #Server CA certificate
// comp-lzo # Use fast LZO compression – may add up to 1 byte per packet for incompressible data.
// 	Enable compression on the VPN link. Don't enable this unless it is also enabled in the server config file.
//====== end config =====
var openVpn = map[string]bool{
	"proto":         true,
	"cipher":        true,
	"ping-restart":  true, // the parameter no exist in default server config OvenVpn
	"ping":          true, // the parameter no exist in default server config OvenVpn
	"connect-retry": true, // the parameter no exist in default server config OvenVpn
	"ca":            true,
	"comp-lzo":      true,

	// A helper directive designed to simplify the expression
	// of --ping and --ping-restart in server mode configurations.
	"keepalive": true,
}

// EndpointMessageTemplate structure for Endpoint message template
type EndpointMessageTemplate struct {
	TemplateHash           string            `json:"template_hash"`
	PaymentReceiverAddress string            `json:"payment_receiver_address"`
	ServiceEndpointAddress string            `json:"service_endpoint_address"`
	Username               string            `json:"username"`
	Password               string            `json:"password"`
	AdditionalParams       map[string]string `json:"additional_params"`
}

// NewEndpointMessageTemplate creates the EndpointMessageTemplate object
// username and password optional fields
func NewEndpointMessageTemplate(
	hash,
	receiver,
	endpoint,
	username,
	password string,
) (*EndpointMessageTemplate, error) {
	if hash == "" || receiver == "" || endpoint == "" {
		return nil, errors.New(ErrInput)
	}

	rWords := strings.Split(receiver, ":")
	if len(rWords) != 2 || rWords[0] == "" || rWords[1] == "" {
		return nil, errors.New(ErrReceiver)
	}

	eWords := strings.Split(endpoint, ":")
	if len(eWords) != 2 || eWords[0] == "" || eWords[1] == "" {
		return nil, errors.New(ErrEndpoint)
	}

	// todo: check hash format [maxim]

	return &EndpointMessageTemplate{
		TemplateHash:           hash,
		PaymentReceiverAddress: receiver,
		ServiceEndpointAddress: endpoint,
		Username:               username,
		Password:               password,
	}, nil
}

// ParsParamsFromConfig function parsing "additional params" from the file
func (msg *EndpointMessageTemplate) ParsParamsFromConfig(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, errors.New(ErrFilePathIsEmpty)
	}
	params, err := util.ParseLinesFromFile(filePath, openVpn, msg.parse)
	if err != nil {
		return nil, errors.Wrap(err, ErrParsingLines)
	}
	ca := params["ca"]
	if ca == "" {
		return nil, errors.New(ErrCertNotExist)
	}
	mainCertPEMBlock, err := ioutil.ReadFile(filepath.Dir(filePath) + string(os.PathSeparator) + ca)
	if err != nil {
		return nil, errors.Wrap(err, ErrCertCanNotRead)
	}
	var cert tls.Certificate
	certPEMBlock := mainCertPEMBlock

	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}

	if len(cert.Certificate) == 0 {
		return nil, errors.New(ErrCertIsNull)
	}

	params["ca"] = string(mainCertPEMBlock)
	msg.AdditionalParams = params
	return json.Marshal(params)
}

func (msg *EndpointMessageTemplate) parse(keys map[string]bool, line string) (string, string, bool) {
	str := strings.TrimSpace(line)
	for key := range keys {
		if strings.HasPrefix(str, key) {
			//find #
			index := strings.Index(str, "#")
			switch index {
			case -1:
				words := strings.Split(str, " ")
				if len(words) == 1 {
					return key, "", true
				}
				value := strings.Join(words[1:], " ")
				return key, value, true
			default:
				subStr := strings.TrimSpace(str[:index])
				words := strings.Split(subStr, " ")
				if len(words) == 1 {
					return key, "", true
				}
				value := strings.Join(words[1:], " ")
				return key, value, true
			}
		}
	}
	return "", "", false
}
