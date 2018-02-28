package main

import (
	"database/sql"
	"flag"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	vpnmon "github.com/privatix/dappctrl/vpn/mon"
	vpnserv "github.com/privatix/dappctrl/vpn/serv"
	"log"
)

//go:generate go generate github.com/privatix/dappctrl/data

type config struct {
	DB         *data.DBConfig
	Log        *util.LogConfig
	VPNServer  *vpnserv.Config
	VPNMonitor *vpnmon.Config
}

func newConfig() *config {
	return &config{
		DB:         data.NewDBConfig(),
		Log:        util.NewLogConfig(),
		VPNServer:  vpnserv.NewConfig(),
		VPNMonitor: vpnmon.NewConfig(),
	}
}

func main() {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	flag.Parse()

	conf := newConfig()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s\n", err)
	}

	db, err := data.NewDB(conf.DB, logger)
	if err != nil {
		logger.Fatal("failed to open DB connection: %s\n", err)
	}
	defer db.DBInterface().(*sql.DB).Close()

	server := vpnserv.NewServer(conf.VPNServer, logger, db)
	go func() {
		logger.Fatal("failed to serve VPN session requests: %s\n",
			server.ListenAndServe())
	}()

	monitor := vpnmon.NewMonitor(conf.VPNMonitor, logger, db)
	go func() {
		logger.Fatal("failed to monitor VPN traffic: %s\n",
			monitor.MonitorTraffic())
	}()

	<-make(chan bool)
}
