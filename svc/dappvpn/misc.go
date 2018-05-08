package main

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"os"
	"path/filepath"
)

const chanPerm = 0644

func commonName() string {
	cn := os.Getenv("common_name")
	if len(cn) == 0 {
		logger.Fatal("empty common_name")
	}
	return base64.URLEncoding.EncodeToString([]byte(cn))
}

func storeChannel(ch string) {
	name := filepath.Join(conf.ChannelDir, commonName())
	err := ioutil.WriteFile(name, []byte(ch), chanPerm)
	if err != nil {
		logger.Fatal("failed to store channel: %s", err)
	}
}

func loadChannel() string {
	name := filepath.Join(conf.ChannelDir, commonName())
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

	if len(os.Args) < 2 {
		logger.Fatal("no filename passed to read credentials")
	}

	file, err := os.Open(os.Args[1])
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
