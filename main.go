package main

import (
	"flag"
	"log"

	"github.com/privatix/dappctrl/agent/uisrv"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/payment"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/util"
)

type config struct {
	AgentServer   *uisrv.Config
	DB            *data.DBConfig
	Job           *job.Config
	Log           *util.LogConfig
	PaymentServer *payment.Config
	SOMC          *somc.Config
}

func newConfig() *config {
	return &config{
		DB:   data.NewDBConfig(),
		Job:  job.NewConfig(),
		Log:  util.NewLogConfig(),
		SOMC: somc.NewConfig(),
	}
}

func main() {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	flag.Parse()

	conf := newConfig()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s", err)
	}

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	db, err := data.NewDB(conf.DB, logger)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db)
	go func() {
		logger.Fatal("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	pmt := payment.NewServer(conf.PaymentServer, logger, db)
	go func() {
		logger.Fatal("failed to start payment server: %s",
			pmt.ListenAndServe())
	}()

	queue := job.NewQueue(conf.Job, logger, db, jobHandlers)
	defer queue.Close()
	logger.Fatal("failed to process job queue: %s", queue.Process())
}
