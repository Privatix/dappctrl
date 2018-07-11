package bugsnag

import (
	"bufio"
	"bytes"

	"github.com/privatix/dappctrl/statik"
)

const (
	pkgFile = "/pkgList/packages.txt"
	mainPkg = "github.com/privatix/dappctrl/main*"
)

func pkgList(excluded map[string]bool) (result []string, err error) {
	data, err := statik.ReadFile(pkgFile)
	if err != nil {
		return
	}

	r := bytes.NewReader(data)

	scanner := bufio.NewScanner(r)

	var pkgs []string

	for scanner.Scan() {
		line := scanner.Text()
		if excluded[line] {
			continue
		}

		pkgs = append(pkgs, line+"*")

	}

	if err = scanner.Err(); err != nil {
		return
	}

	result = append(result, mainPkg)

	// all except the parent package
	result = append(result, pkgs[1:]...)
	return
}
