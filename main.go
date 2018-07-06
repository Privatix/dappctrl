package main

import (
	"flag"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	vpncli "github.com/privatix/dappctrl/messages/ept/config"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/handlers"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/uisrv"
	"github.com/privatix/dappctrl/util"
)

type ethConfig struct {
	Contract struct {
		PTCAddrHex string
		PSCAddrHex string
	}
	GethURL string
}

type config struct {
	AgentServer   *uisrv.Config
	BlockMonitor  *monitor.Config
	ClientMonitor *cbill.Config
	Eth           *ethConfig
	DB            *data.DBConfig
	Gas           *worker.GasConf
	Job           *job.Config
	Log           *util.LogConfig
	PayServer     *pay.Config
	PayAddress    string
	Proc          *proc.Config
	Report        *bugsnag.Config
	ServiceRunner *svcrun.Config
	SessionServer *sesssrv.Config
	SOMC          *somc.Config
	StaticPasword string
	VPNClient     *vpncli.Config
}

func newConfig() *config {
	return &config{
		BlockMonitor:  monitor.NewConfig(),
		ClientMonitor: cbill.NewConfig(),
		DB:            data.NewDBConfig(),
		AgentServer:   uisrv.NewConfig(),
		Job:           job.NewConfig(),
		Log:           util.NewLogConfig(),
		Proc:          proc.NewConfig(),
		Report:        bugsnag.NewConfig(),
		ServiceRunner: svcrun.NewConfig(),
		SessionServer: sesssrv.NewConfig(),
		SOMC:          somc.NewConfig(),
		VPNClient:     vpncli.NewConfig(),
	}
}

func readConfig(conf *config) {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	flag.Parse()
	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		log.Fatalf("failed to read configuration: %s", err)
	}
}

func getPWDStorage(conf *config) data.PWDGetSetter {
	if conf.StaticPasword == "" {
		return new(data.PWDStorage)
	}
	storage := data.StaticPWDStorage(conf.StaticPasword)
	return &storage
}

func main() {
	defer bugsnag.PanicHunter()

	conf := newConfig()
	readConfig(conf)

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}
	defer logger.GracefulStop()

	db, err := data.NewDB(conf.DB, logger)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	reporter := bugsnag.NewClient(conf.Report, db, logger)
	logger.Reporter(reporter)

	gethConn, err := ethclient.Dial(conf.Eth.GethURL)
	if err != nil {
		logger.Fatal("failed to dial geth node: %v", err)
	}

	ptcAddr := common.HexToAddress(conf.Eth.Contract.PTCAddrHex)
	ptc, err := contract.NewPrivatixTokenContract(ptcAddr, gethConn)
	if err != nil {
		logger.Fatal("failed to create ptc instance: %v", err)
	}

	pscAddr := common.HexToAddress(conf.Eth.Contract.PSCAddrHex)

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

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		panic(err)
	}

	pwdStorage := getPWDStorage(conf)

	worker, err := worker.NewWorker(logger, db, somcConn,
		worker.NewEthBackend(psc, ptc, gethConn), conf.Gas,
		pscAddr, conf.PayAddress, pwdStorage, data.ToPrivateKey,
		conf.VPNClient)
	if err != nil {
		panic(err)
	}

	queue := job.NewQueue(conf.Job, logger, db, handlers.HandlersMap(worker))
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, queue)
	worker.SetProcessor(pr)

	runner := svcrun.NewServiceRunner(conf.ServiceRunner, logger, db, pr)
	worker.SetRunner(runner)
	worker.ManagementInterfacePort(
		managementPortForClient(conf.ServiceRunner, logger, db.Querier))

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, queue, pwdStorage, pr)

	go func() {
		logger.Fatal("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	amon, err := abill.NewMonitor(
		time.Duration(5)*time.Second, db, logger, pr)
	if err != nil {
		logger.Fatal("failed to create agent billing monitor: %s", err)
	}
	go func() {
		logger.Fatal("failed to run agent billing monitor: %s",
			amon.Run())
	}()

	cmon := cbill.NewMonitor(conf.ClientMonitor,
		logger, db, pr, conf.Eth.Contract.PSCAddrHex, pwdStorage)
	go func() {
		logger.Fatal("failed to run client billing monitor: %s",
			cmon.Run())
	}()

	mon, err := monitor.NewMonitor(conf.BlockMonitor, logger, db, queue,
		gethConn, pscAddr, ptcAddr)

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
