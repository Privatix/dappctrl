package msg

import (
	"io/ioutil"
	"os"

	"github.com/privatix/dappctrl/statik"
)

const (
	pathPerm = 0755
)

func notExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return true
	}
	return false
}

func checkFile(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func writeFile(name string, data []byte) error {
	return ioutil.WriteFile(name, data, filePerm)
}

func readFileFromVirtualFS(name string) ([]byte, error) {
	return statik.ReadFile(name)
}

func makeDir(name string) error {
	return os.MkdirAll(name, pathPerm)
}
