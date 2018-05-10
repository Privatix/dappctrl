package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// IsUUID checks if a given string is a UUID.
func IsUUID(s string) bool {
	_, err := uuid.FromString(s)
	return err == nil
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

// RootPath returns a path of the root package.
func RootPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "?"
	}
	return filepath.Dir(filepath.Dir(file))
}

// Caller returns a caller's call location. If F1 calls F2 which in its turn
// calls Caller, then this function will return a location within F1 where it
// calls F2.
func Caller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "?"
	}

	rel, err := filepath.Rel(RootPath(), file)
	if err != nil {
		return "?"
	}

	return fmt.Sprintf("%s:%d", rel, line)
}
