package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	chanPerm = 0644

	jsonIdent = "    "
)

func encode(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

func commonNameOrEmpty() string {
	return os.Getenv("common_name")
}

func commonName() string {
	cn := commonNameOrEmpty()
	if len(cn) == 0 {
		logger.Fatal("empty common_name")
	}
	return cn
}

func storeChannel(cn, ch string) {
	name := filepath.Join(conf.ChannelDir, encode(cn))
	err := ioutil.WriteFile(name, []byte(ch), chanPerm)
	if err != nil {
		logger.Fatal("failed to store channel: %s", err)
	}
}

func loadChannel() string {
	name := filepath.Join(conf.ChannelDir, encode(commonName()))
	data, err := ioutil.ReadFile(name)
	if err != nil {
		logger.Fatal("failed to load channel: %s", err)
	}
	return string(data)
}

func getCreds() (string, string) {
	user := os.Getenv("username")
	pass := os.Getenv("password")

	if len(user) != 0 && len(pass) != 0 {
		return user, pass
	}

	if flag.NArg() < 1 {
		logger.Fatal("no filename passed to read credentials")
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		logger.Fatal("failed to open file with credentials: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	user = scanner.Text()
	scanner.Scan()
	pass = scanner.Text()

	if err := scanner.Err(); err != nil {
		logger.Fatal("failed to read file with credentials: %s", err)
	}

	return user, pass
}
