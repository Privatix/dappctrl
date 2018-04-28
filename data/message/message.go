package message

import (
	"github.com/pkg/errors"
	"github.com/privatix/dappctrl/util"
	"strings"
)

var openVpn = map[string]bool{
	"proto":         true,
	"cipher":        true,
	"ping-restart":  true,
	"ping":          true,
	"connect-retry": true,
	"ca":            true,
	"comp-lzo":      true,
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
		return nil, errors.New("one or more input parameters is wrong")
	}
	return &EndpointMessageTemplate{
		TemplateHash:           hash,
		PaymentReceiverAddress: receiver,
		ServiceEndpointAddress: endpoint,
		Username:               username,
		Password:               password,
	}, nil
}

func (msg *EndpointMessageTemplate) ParsParamsFromConfig(filePath string) (map[string]string, error) {
	return util.ParseLinesFromFile(filePath, openVpn, msg.parse)
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
