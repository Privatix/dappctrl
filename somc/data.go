package somc

import (
	"strings"

	"github.com/privatix/dappctrl/data"
)

var separator = "&" // Character not present in url base64.

func combineURLBase64Strings(args []string) string {
	return strings.Join(args, separator)
}

func extractURLBase64Strings(somcData string) []data.Base64String {
	ret := []data.Base64String{}
	parts := strings.Split(somcData, separator)
	for _, part := range parts {
		ret = append(ret, data.Base64String(part))
	}
	return ret
}
