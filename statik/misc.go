package statik

import (
	"fmt"
	"io/ioutil"

	"github.com/rakyll/statik/fs"
)

//go:generate rm -f statik.go
//go:generate statik -f -src=. -dest=..

// File paths.
const (
	EndpointJSONSchema = "/templates/ept.json"
)

// ReadFile reads a file content from the embedded filesystem.
func ReadFile(name string) ([]byte, error) {
	fs, err := fs.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open statik FS: %s", err)
	}

	file, err := fs.Open(EndpointJSONSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to open statik file: %s", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read statik file: %s", err)
	}

	return data, nil
}
