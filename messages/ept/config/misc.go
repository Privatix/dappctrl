package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util"
)

// ParseCertFromFile Parsing TLS certificate from file
func ParseCertFromFile(caCertPath string) (string, error) {
	mainCertPEMBlock, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return "", ErrCertCanNotRead
	}

	if !util.IsTLSCert(string(mainCertPEMBlock)) {
		return "", ErrCertNotFound
	}

	return string(mainCertPEMBlock), nil
}

func isHost(host string) bool {
	if util.IsHostname(host) || util.IsIPv4(host) {
		return true
	}
	return false
}

func push(ctx context.Context, errC chan error, req *pushReq) {
	err := sesssrv.Post(req.sessSrvConfig,
		req.username, req.password, sesssrv.PathProductConfig,
		req.args, nil)

	select {
	case <-ctx.Done():
		return
	case errC <- err:
	}
}

func productArgs(confPath, caPath string,
	keys []string) (sesssrv.ProductArgs, error) {
	result := sesssrv.ProductArgs{}

	conf, err := ServerConfig(confPath, false, keys)
	if err != nil {
		return result, err
	}

	ca, err := ParseCertFromFile(caPath)
	if err != nil {
		return result, err
	}

	conf[caData] = ca

	result.Config = conf

	return result, nil
}

func delDup(keys []string) map[string]bool {
	// delete duplicates
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[strings.TrimSpace(key)] = true
	}
	return keyMap
}

func searchCa(paths []string) (string, string, bool) {
	for _, filePath := range paths {
		cert, err := ParseCertFromFile(filePath)
		if err == nil {
			return cert, filePath, true
		}
	}

	return "", "", false
}

func fillCa(results map[string]string, filePath string) error {
	// check ca key
	ca := results[caNameFromConfig]
	if ca == "" {
		return ErrCertNotExist
	}

	// absolute path
	absPath := filepath.Join(filepath.Dir(filePath), ca)

	cert, certPath, found := searchCa([]string{ca, absPath})
	if !found {
		return ErrCertNotFound
	}

	results[caData] = cert
	results[caPathName] = certPath

	return nil
}

func createPath(target string) error {
	return os.MkdirAll(target, pathPerm)
}

func notExist(location string) bool {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		return true
	}
	return false
}
