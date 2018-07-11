package msg

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/privatix/dappctrl/util"
)

func vpnParams(file string,
	keys []string) (params map[string]string, err error) {
	params = make(map[string]string)

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read vpn"+
			" configuration file")
	}
	defer f.Close()

	keyMap := make(map[string]bool)

	for _, key := range keys {
		keyMap[strings.TrimSpace(key)] = true
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if key, value, add :=
			parseLine(keyMap, scanner.Text()); add {
			if key != "" {
				params[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read vpn"+
			" configuration file")
	}

	return params, err
}

func certificateAuthority(file string) (ca []byte, err error) {
	mainCertPEMBlock, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read Certificate"+
			" Authority file")
	}

	if !util.IsTLSCert(string(mainCertPEMBlock)) {
		return nil, errors.Wrap(err, "certificate authority"+
			" can not be found in the specified path")
	}

	return mainCertPEMBlock, nil
}

func parseLine(keys map[string]bool, line string) (string, string, bool) {
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
