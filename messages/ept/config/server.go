package config

import (
	"bufio"
	"context"
	"os"
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

type pushReq struct {
	sessSrvConfig *srv.Config
	username      string
	password      string
	args          sesssrv.ProductArgs
}

// PushConfigReq the parameters that are needed
// to send the configuration to the server
type PushConfigReq struct {
	username string
	password string
	confPath string
	caPath   string
	keys     []string
	retrySec int64
}

// NewPushConfigReq fills the request structure
func NewPushConfigReq(username, password, confPath,
	caPath string, keys []string, retrySec int64) *PushConfigReq {
	return &PushConfigReq{
		username: username,
		password: password,
		confPath: confPath,
		caPath:   caPath,
		keys:     keys,
		retrySec: retrySec,
	}
}

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

	keyMap := delDup(keys)

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

	if withCa {
		if err := fillCa(results, filePath); err != nil {
			return nil, err
		}
	}

	return results, nil
}

func parseLine(keys map[string]bool,
	line string) (string, string, bool) {
	str := strings.TrimSpace(line)

	for key := range keys {
		sStr := strings.Split(str, " ")
		if sStr[0] != key {
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
	logger *util.Logger, in *PushConfigReq) error {
	if in.retrySec <= 0 || logger == nil || sessSrvConfig == nil {
		return ErrInput
	}

	args, err := productArgs(in.confPath, in.caPath, in.keys)
	if err != nil {
		return err
	}

	req := &pushReq{
		sessSrvConfig: sessSrvConfig,
		username:      in.username,
		password:      in.password, args: args,
	}

	errC := make(chan error)

	pushed := false

	for {
		if pushed {
			break
		}
		go push(ctx, errC, req)
		select {
		case <-ctx.Done():
			return ErrCancel
		case err := <-errC:
			if err != nil {
				logger.Warn("Failed to push app config to"+
					" dappctrl. Original error: %s", err)
				time.Sleep(time.Second *
					time.Duration(in.retrySec))
				continue
			}
			pushed = true
		}

	}

	return nil
}
