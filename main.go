package main

import (
	"flag"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/eth/truffle"
	"github.com/privatix/dappctrl/execsrv"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/uisrv"
	"github.com/privatix/dappctrl/util"
)

type ethConfig struct {
	Contract struct {
		PTCAddr string
		PSCAddr string
	}
	GethURL       string
	TruffleAPIURL string
}

type config struct {
	AgentServer   *uisrv.Config
	BlockMonitor  *monitor.Config
	Eth           *ethConfig
	DB            *data.DBConfig
	Job           *job.Config
	Log           *util.LogConfig
	PayServer     *pay.Config
	PayAddress    string
	Proc          *proc.Config
	SessionServer *sesssrv.Config
	SOMC          *somc.Config
}

func newConfig() *config {
	return &config{
		BlockMonitor:  monitor.NewConfig(),
		DB:            data.NewDBConfig(),
		AgentServer:   uisrv.NewConfig(),
		Job:           job.NewConfig(),
		Log:           util.NewLogConfig(),
		Proc:          proc.NewConfig(),
		SessionServer: sesssrv.NewConfig(),
		SOMC:          somc.NewConfig(),
	}
}

func readConfig(conf *config) {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	flag.Parse()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s", err)
	}
	// If test truffle api is specified, pull and update contract addresses.
	if conf.Eth.TruffleAPIURL != "" {
		api := truffle.API(conf.Eth.TruffleAPIURL)
		conf.Eth.Contract.PSCAddr = api.FetchPSCAddress()
		conf.Eth.Contract.PTCAddr = api.FetchPTCAddress()
	}
}

func main() {
	conf := newConfig()
	readConfig(conf)

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	db, err := data.NewDB(conf.DB, logger)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	gethConn, err := ethclient.Dial(conf.Eth.GethURL)
	if err != nil {
		logger.Fatal("failed to dial geth node: %v", err)
	}

	ptcAddr := common.BytesToAddress([]byte(conf.Eth.Contract.PTCAddr))
	ptc, err := contract.NewPrivatixTokenContract(ptcAddr, gethConn)
	if err != nil {
		logger.Fatal("failed to create ptc instance: %v", err)
	}

	pscAddr := common.BytesToAddress([]byte(conf.Eth.Contract.PSCAddr))

	psc, err := contract.NewPrivatixServiceContract(pscAddr, gethConn)
	if err != nil {
		logger.Fatal("failed to create psc intance: %v", err)
	}

	paySrv := pay.NewServer(conf.PayServer, logger, db)
	go func() {
		logger.Fatal("failed to start pay server: %s",
			paySrv.ListenAndServe())
	}()

	sess := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		logger.Fatal("failed to start session server: %s",
			sess.ListenAndServe())
	}()

	// TODO: Remove when not needed anymore.
	exec := execsrv.NewServer(logger)
	go func() {
		logger.Fatal("failed to start exec server: %s",
			exec.ListenAndServe())
	}()

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		panic(err)
	}

	pwdStorage := new(data.PWDStorage)

	handler, err := worker.NewWorker(logger, db, somcConn,
		worker.NewEthBackend(psc, ptc, gethConn),
		pscAddr, conf.PayAddress, pwdStorage, data.ToPrivateKey)
	if err != nil {
		panic(err)
	}

	queue := job.NewQueue(conf.Job, logger, db, proc.HandlersMap(handler))
	handler.SetQueue(queue)

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, queue, pwdStorage)

	go func() {
		logger.Fatal("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	mon, err := monitor.NewMonitor(conf.BlockMonitor, logger, db, queue,
		gethConn, pscAddr)
	if err != nil {
		logger.Fatal("failed to initialize"+
			" the blockchain monitor: %v", err)
	}

	if err := mon.Start(); err != nil {
		logger.Fatal("failed to start"+
			" the blockchain monitor: %v", err)
	}
	defer mon.Stop()

	logger.Fatal("failed to process job queue: %s", queue.Process())
}
