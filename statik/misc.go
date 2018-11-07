package statik

import (
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/rakyll/statik/fs"
)

// File paths.
const (
	EndpointJSONSchema = "/templates/ept.json"
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

// OpenFile opens the named file from the embedded filesystem.
func OpenFile(name string) (http.File, error) {
	fs, err := fs.New()
	if err != nil {
		return nil, ErrOpenFS
	}

	file, err := fs.Open(name)
	if err != nil {
		return nil, ErrOpenFile
	}
	return file, nil
}

// ReadDir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values.
func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := OpenFile(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, nil
}
