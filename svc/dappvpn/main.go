package main

import (
	"log"
	"os"
	"strconv"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

type config struct {
	ChannelDir string // Directory for common-name -> channel mappings.
	Log        *util.LogConfig
	Monitor    *mon.Config
	Password   string // HTTP basic auth. password.
	Product    string // HTTP basic auth. username.
	Server     *srv.Config
}

func newConfig() *config {
	return &config{
		ChannelDir: ".",
		Log:        util.NewLogConfig(),
		Monitor:    mon.NewConfig(),
		Server:     srv.NewConfig(),
	}
}

var (
	conf   *config
	logger *util.Logger
)

func main() {
	var err error

	conf = newConfig()
	name := util.ExeDirJoin("dappvpn.config.json")
	if err := util.ReadJSONFile(name, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}

	logger, err = util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s\n", err)
	}

	switch os.Getenv("script_type") {
	case "user-pass-verify":
		handleAuth()
	case "client-connect":
		handleConnect()
	case "client-disconnect":
		handleDisconnect()
	default:
		handleMonitor()
	}
}

func handleAuth() {
	user, pass := getCreds()
	args := sesssrv.AuthArgs{ClientID: user, Password: pass}

	err := sesssrv.Post(conf.Server,
		conf.Product, conf.Password, sesssrv.PathAuth, args, nil)
	if err != nil {
		logger.Fatal("failed to auth: %s", err)
	}

	storeChannel(user)
}

func handleConnect() {
	port, err := strconv.Atoi(os.Getenv("trusted_port"))
	if err != nil || port <= 0 || port > 0xFFFF {
		logger.Fatal("bad trusted_port value")
	}

	args := sesssrv.StartArgs{
		ClientID:   loadChannel(),
		ClientIP:   os.Getenv("trusted_ip"),
		ClientPort: uint16(port),
	}

	err = sesssrv.Post(conf.Server,
		conf.Product, conf.Password, sesssrv.PathStart, args, nil)
	if err != nil {
		logger.Fatal("failed to start session: %s", err)
	}
}

func handleDisconnect() {
	down, err := strconv.ParseUint(os.Getenv("bytes_sent"), 10, 64)
	if err != nil || down < 0 {
		log.Fatalf("bad bytes_sent value")
	}

	up, err := strconv.ParseUint(os.Getenv("bytes_received"), 10, 64)
	if err != nil || up < 0 {
		log.Fatalf("bad bytes_received value")
	}

	args := sesssrv.StopArgs{
		ClientID: loadChannel(),
		Units:    down + up,
	}

	err = sesssrv.Post(conf.Server,
		conf.Product, conf.Password, sesssrv.PathStop, args, nil)
	if err != nil {
		logger.Fatal("failed to stop session: %s", err)
	}
}

func handleMonitor() {
	handleByteCount := func(ch string, up, down uint64) bool {
		args := sesssrv.UpdateArgs{
			ClientID: ch,
			Units:    down + up,
		}

		err := sesssrv.Post(conf.Server, conf.Product,
			conf.Password, sesssrv.PathUpdate, args, nil)

		if err != nil {
			logger.Info("failed to update session %s: %s", ch, err)
			return false
		}

		return true
	}

	monitor := mon.NewMonitor(conf.Monitor, logger, handleByteCount)

	logger.Fatal("failed to monitor vpn traffic: %s",
		monitor.MonitorTraffic())
}