package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"

	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
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

type config struct {
	AgentServer   *uisrv.Config
	BlockMonitor  *monitor.Config
	ClientMonitor *cbill.Config
	EptMsg        *ept.Config
	Eth           *eth.Config
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
		EptMsg:        ept.NewConfig(),
		Eth:           eth.NewConfig(),
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
	fatal := make(chan string)
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

	reporter, err := bugsnag.NewClient(conf.Report, db, logger)
	if err != nil {
		logger.Fatal("failed to create Bugsnag client: %s", err)
	}

	logger.Reporter(reporter)

	gethConn, err := eth.NewEtherClient(conf.Eth)
	if err != nil {
		logger.Fatal("failed to dial geth node: %v", err)
	}
	defer gethConn.Close()

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
		fatal <- fmt.Sprintf("failed to start pay server: %s",
			paySrv.ListenAndServe())
	}()
	defer paySrv.Close()

	sess := sesssrv.NewServer(conf.SessionServer, logger, db)
	go func() {
		fatal <- fmt.Sprintf("failed to start session server: %s",
			sess.ListenAndServe())
	}()
	defer sess.Close()

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		logger.Fatal("failed to connect to SOMC: %s", err)
	}
	defer somcConn.Close()

	pwdStorage := getPWDStorage(conf)

	worker, err := worker.NewWorker(logger, db, somcConn,
		worker.NewEthBackend(psc, ptc, gethConn), conf.Gas,
		pscAddr, conf.PayAddress, pwdStorage, data.ToPrivateKey,
		conf.VPNClient, conf.EptMsg, conf.Eth)
	if err != nil {
		logger.Fatal("failed to create worker: %s", err)
	}

	queue := job.NewQueue(conf.Job, logger, db, handlers.HandlersMap(worker))
	defer queue.Close()
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, queue)
	worker.SetProcessor(pr)

	runner := svcrun.NewServiceRunner(conf.ServiceRunner, logger, db, pr)
	defer runner.StopAll()
	worker.SetRunner(runner)

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, queue, pwdStorage, pr)
	go func() {
		fatal <- fmt.Sprintf("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	amon, err := abill.NewMonitor(
		time.Duration(5)*time.Second, db, logger, pr)
	if err != nil {
		logger.Fatal("failed to create agent billing monitor: %s", err)
	}
	go func() {
		fatal <- fmt.Sprintf("failed to run agent billing monitor: %s",
			amon.Run())
	}()

	cmon := cbill.NewMonitor(conf.ClientMonitor,
		logger, db, pr, conf.Eth.Contract.PSCAddrHex, pwdStorage)
	go func() {
		fatal <- fmt.Sprintf("failed to run client billing monitor: %s",
			cmon.Run())
	}()
	defer cmon.Close()

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

	go func() {
		fatal <- fmt.Sprintf("failed to process job queue: %s",
			queue.Process())
	}()

	logger.Fatal(<-fatal)
}
