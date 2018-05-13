package config

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

const (
	caNameFromConfig = "ca"
	caPathName       = "caPathName"
	caData           = "caData"
)

// ServerConfig parsing OpenVpn config file and parsing
// certificate from file.
func ServerConfig(filePath string, withCa bool,
	keys []string) (map[string]string, error) {
	if filePath == "" {
		return nil, ErrFilePathIsEmpty
	}
	return parseConfig(filePath, keys, withCa)
}

func parseConfig(filePath string,
	keys []string, withCa bool) (map[string]string, error) {
	// check input
	if keys == nil || filePath == "" {
		return nil, ErrInput
	}

	// open config file
	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	// delete duplicates
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[strings.TrimSpace(key)] = true
	}

	results := make(map[string]string)

	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		if key, value, add :=
			parseLine(keyMap, scanner.Text()); add {
			if key != "" {
				results[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// A certificate can be located on an absolute and relative path.
	// If the certificate is found in the file,
	// will return the body of the certificate,
	// the absolute path to the certificate, and true
	pCert := func(paths []string) (string, string, bool) {
		for _, filePath := range paths {
			cert, err := ParseCertFromFile(filePath)
			if err == nil {
				return cert, filePath, true
			}
		}

		return "", "", false
	}

	// function is thread-safe, mutex is not required
	findCa := func() error {
		// check ca key
		ca := results[caNameFromConfig]
		if ca == "" {
			return ErrCertNotExist
		}

		// absolute path
		absPath := filepath.Dir(filePath) +
			string(os.PathSeparator) + ca

		cert, certPath, found := pCert([]string{ca, absPath})
		if !found {
			return ErrCertNotFound
		}

		results[caData] = cert
		results[caPathName] = certPath

		return nil
	}

	if withCa {
		if err := findCa(); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func parseLine(keys map[string]bool,
	line string) (string, string, bool) {
	str := strings.TrimSpace(line)

	for key := range keys {
		if !strings.HasPrefix(str, key) {
			continue
		}

		index := strings.Index(str, "#")

		if index == -1 {
			words := strings.Split(str, " ")
			if len(words) == 1 {
				return key, "", true
			}
			value := strings.Join(words[1:], " ")
			return key, value, true
		}

		subStr := strings.TrimSpace(str[:index])

		words := strings.Split(subStr, " ")

		if len(words) == 1 {
			return key, "", true
		}

		value := strings.Join(words[1:], " ")

		return key, value, true
	}
	return "", "", false
}

// PushConfig push OpenVpn config to Session Server
// Hostname or ip address and port from Session Server
// is taken from sessSrvConfig. Function can be canceled via context.
// username and password needed for access the product in the database.
// The confPath, caPath are absolute paths to OpenVpn config and Ca files,
// In variable keys the list of keys for parsing.
// The timeout in seconds between attempts to send data to the server must
// be specified in variable retrySec
func PushConfig(ctx context.Context, sessSrvConfig *srv.Config,
	logger *util.Logger, username, password, confPath,
	caPath string, keys []string, retrySec int64) error {
	if retrySec <= 0 {
		return ErrRetrySec
	}
	conf, err := ServerConfig(confPath, false, keys)
	if err != nil {
		return err
	}

	ca, err := ParseCertFromFile(caPath)
	if err != nil {
		return err
	}

	conf[caData] = ca

	args := sesssrv.ProductArgs{
		Config: conf,
	}

	errC := make(chan error)

	push := func() {
		err := sesssrv.Post(sessSrvConfig,
			username, password, sesssrv.PathProductConfig,
			args, nil)

		select {
		case <-ctx.Done():
			return
		case errC <- err:
		}
	}

	for {
		go push()
		select {
		case <-ctx.Done():
			return ErrCancel
		case err := <-errC:
			if err != nil {
				logger.Warn("Failed to push app config to"+
					" dappctrl. Original error: %s", err)
				time.Sleep(time.Second * time.Duration(retrySec))
				continue
			}
			return nil
		}

	}

	return nil
}
