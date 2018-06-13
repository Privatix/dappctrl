package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/svc/dappvpn/mon"
	"github.com/privatix/dappctrl/svc/dappvpn/pusher"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

type serverConfig struct {
	*srv.Config
	Password string
	Username string
}

type config struct {
	ChannelDir string // Directory for common-name -> channel mappings.
	Log        *util.LogConfig
	Monitor    *mon.Config
	Pusher     *pusher.Config
	Server     *serverConfig
}

func newConfig() *config {
	return &config{
		ChannelDir: ".",
		Log:        util.NewLogConfig(),
		Monitor:    mon.NewConfig(),
		Pusher:     pusher.NewConfig(),
		Server:     &serverConfig{Config: srv.NewConfig()},
	}
}

var (
	conf   *config
	logger *util.Logger
)

func main() {
	fconfig := flag.String(
		"config", "dappvpn.config.json", "Configuration file")
	flag.Parse()

	conf = newConfig()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}

	var err error
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
		handleMonitor(*fconfig)
	}
}

func handleAuth() {
	user, pass := getCreds()
	args := sesssrv.AuthArgs{ClientID: user, Password: pass}

	err := sesssrv.Post(conf.Server.Config, conf.Server.Username,
		conf.Server.Password, sesssrv.PathAuth, args, nil)
	if err != nil {
		logger.Fatal("failed to auth: %s", err)
	}

	if cn := commonNameOrEmpty(); len(cn) != 0 {
		storeChannel(cn, user)
	}
	storeChannel(user, user) // Needed when using username-as-common-name.
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

	err = sesssrv.Post(conf.Server.Config, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStart, args, nil)
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

	err = sesssrv.Post(conf.Server.Config, conf.Server.Username,
		conf.Server.Password, sesssrv.PathStop, args, nil)
	if err != nil {
		logger.Fatal("failed to stop session: %s", err)
	}
}

func handleMonitor(confFile string) {
	handleByteCount := func(ch string, up, down uint64) bool {
		args := sesssrv.UpdateArgs{
			ClientID: ch,
			Units:    down + up,
		}

		err := sesssrv.Post(conf.Server.Config, conf.Server.Username,
			conf.Server.Password, sesssrv.PathUpdate, args, nil)

		if err != nil {
			msg := "failed to update session for channel %s: %s"
			logger.Info(msg, ch, err)
			return false
		}

		return true
	}

	if !conf.Pusher.Pushed {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			c := pusher.NewCollect(conf.Pusher, conf.Server.Config,
				conf.Server.Username, conf.Server.Password,
				logger)

			if err := pusher.PushConfig(ctx, c); err != nil {
				logger.Error("failed to send OpenVpn"+
					" server configuration: %s\n", err)
			}
		}()
	}

	monitor := mon.NewMonitor(conf.Monitor, logger, handleByteCount)

	logger.Fatal("failed to monitor vpn traffic: %s",
		monitor.MonitorTraffic())
}
