package util

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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
	return uuid.New().String()
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
