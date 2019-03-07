package statik

import (
	"io/ioutil"

	"github.com/rakyll/statik/fs"
)

// ReadFile reads a file content from the embedded filesystem.
func ReadFile(name string) ([]byte, error) {
	fs, err := fs.New()
	if err != nil {
		return nil, ErrOpenFS
	}

	file, err := fs.Open(name)
	if err != nil {
		return nil, ErrOpenFile
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, ErrReadFile
	}

	return data, nil
}
