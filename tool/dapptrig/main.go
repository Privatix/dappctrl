package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/privatix/dappctrl/util"
	vpnsrv "github.com/privatix/dappctrl/vpn/srv"
)

type config struct {
	LogFile    string
	ServerAddr string
	ServerTLS  bool
	ChannelDir string
}

func newConfig() *config {
	conf := vpnsrv.NewConfig()
	return &config{
		LogFile:    "dapptrig.log",
		ServerAddr: conf.Addr,
		ServerTLS:  conf.TLS != nil,
		ChannelDir: "channels",
	}
}

const (
	logPerm  = 0644
	chanPerm = 0644
)

func main() {
	conf := newConfig()
	name := util.ExeDirJoin("dapptrig.config.json")
	if err := util.ReadJSONFile(name, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}

	if len(conf.LogFile) != 0 {
		out, err := os.OpenFile(conf.LogFile,
			os.O_RDWR|os.O_CREATE|os.O_APPEND, logPerm)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer out.Close()

		log.SetOutput(out)
	}

	switch os.Getenv("script_type") {
	case "user-pass-verify":
		handleAuth(conf)
	case "client-connect":
		handleConnect(conf)
	case "client-disconnect":
		handleDisconnect(conf)
	default:
		log.Fatalf("unsupported script_type")
	}
}

func getCreds() (string, string) {
	user := os.Getenv("username")
	pass := os.Getenv("password")

	if len(user) != 0 && len(pass) != 0 {
		return user, pass
	}

	if len(os.Args) < 2 {
		log.Fatalf("no filename passed to read credentials")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal("failed to open file with credentials: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	user = scanner.Text()
	scanner.Scan()
	pass = scanner.Text()

	if err := scanner.Err(); err != nil {
		log.Fatal("failed to read file with credentials: %s", err)
	}

	return user, pass
}

func post(conf *config, path string, req interface{}, rep interface{}) {
	data, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("failed to encode request: %s", err)
	}

	var proto = "http"
	if conf.ServerTLS {
		proto += "s"
	}

	resp, err := http.Post(proto+"://"+conf.ServerAddr+path,
		"application/json", bytes.NewReader(data))
	if err != nil {
		log.Fatalf("failed to post request: %s", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(rep); err != nil {
		log.Fatalf("failed to decode reply: %s", err)
	}
}

func commonName() string {
	cn := os.Getenv("common_name")
	if len(cn) == 0 {
		log.Fatalf("empty common_name")
	}
	return base64.URLEncoding.EncodeToString([]byte(cn))
}

func storeChannel(conf *config, ch string) {
	name := filepath.Join(conf.ChannelDir, commonName())
	err := ioutil.WriteFile(name, []byte(ch), chanPerm)
	if err != nil {
		log.Fatalf("failed to store channel: %s", err)
	}
}

func loadChannel(conf *config) string {
	name := filepath.Join(conf.ChannelDir, commonName())
	data, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatalf("failed to load channel: %s", err)
	}
	return string(data)
}

func handleAuth(conf *config) {
	user, pass := getCreds()

	req := vpnsrv.AuthRequest{Channel: user, Password: pass}

	var rep vpnsrv.AuthReply
	post(conf, vpnsrv.PathAuth, req, &rep)
	if len(rep.Error) != 0 {
		log.Fatalf("failed to authenticate %s: %s", user, rep.Error)
	}

	storeChannel(conf, user)
}

func handleConnect(conf *config) {
	port, err := strconv.Atoi(os.Getenv("trusted_port"))
	if err != nil || port <= 0 || port > 0xFFFF {
		log.Fatalf("bad trusted_port value")
	}

	req := vpnsrv.StartRequest{
		Channel:    loadChannel(conf),
		ServerIP:   os.Getenv("ifconfig_remote"),
		ClientIP:   os.Getenv("trusted_ip"),
		ClientPort: uint16(port),
	}

	var rep vpnsrv.StartReply
	post(conf, vpnsrv.PathStart, req, &rep)
	if len(rep.Error) != 0 {
		log.Fatalf("failed to start session: %s", rep.Error)
	}
}

func handleDisconnect(conf *config) {
	down, err := strconv.Atoi(os.Getenv("bytes_sent"))
	if err != nil || down < 0 {
		log.Fatalf("bad bytes_sent value")
	}

	up, err := strconv.Atoi(os.Getenv("bytes_received"))
	if err != nil || up < 0 {
		log.Fatalf("bad bytes_received value")
	}

	req := vpnsrv.StopRequest{
		Channel:    loadChannel(conf),
		Uploaded:   uint64(up),
		Downloaded: uint64(down),
	}

	var rep vpnsrv.StopReply
	post(conf, vpnsrv.PathStop, req, &rep)
	if len(rep.Error) != 0 {
		log.Fatalf("failed to stop session: %s", rep.Error)
	}
}
