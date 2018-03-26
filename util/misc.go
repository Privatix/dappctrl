package util

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common/number"
	uuid "github.com/satori/go.uuid"
)

// ReadJSONFile reads and parses a JSON file filling a given data instance.
func ReadJSONFile(name string, data interface{}) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(data)
}

// NewUUID generates a new UUID.
func NewUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

// ExeDirJoin composes a file name relative to a running executable.
func ExeDirJoin(elem ...string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	elem = append([]string{filepath.Dir(exe)}, elem...)
	return filepath.Join(elem...)
}

// Base64ToEthNum returns eth's number.Number from base64 encoded string.
func Base64ToEthNum(b64X string) (*number.Number, error) {
	b, err := base64.URLEncoding.DecodeString(strings.TrimSpace(b64X))
	if err != nil {
		return nil, err
	}
	x := number.Big(0)
	x.SetBytes(b)
	return x, nil
}
