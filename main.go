package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	dblog "github.com/privatix/dappctrl/data/log"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/handlers"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/uisrv"
	"github.com/privatix/dappctrl/util"
	log2 "github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

type config struct {
	AgentMonitor  *abill.Config
	AgentServer   *uisrv.Config
	BlockMonitor  *monitor.Config
	ClientMonitor *cbill.Config
	DB            *data.DBConfig
	DBLog         *dblog.Config
	EptMsg        *ept.Config
	Eth           *eth.Config
	FileLog       *log2.FileConfig
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
	UI            *ui.Config
}

func newConfig() *config {
	return &config{
		AgentMonitor:  abill.NewConfig(),
		AgentServer:   uisrv.NewConfig(),
		BlockMonitor:  monitor.NewConfig(),
		ClientMonitor: cbill.NewConfig(),
		DB:            data.NewDBConfig(),
		DBLog:         dblog.NewConfig(),
		EptMsg:        ept.NewConfig(),
		Eth:           eth.NewConfig(),
		FileLog:       log2.NewFileConfig(),
		Job:           job.NewConfig(),
		Log:           util.NewLogConfig(),
		Proc:          proc.NewConfig(),
		Report:        bugsnag.NewConfig(),
		ServiceRunner: svcrun.NewConfig(),
		SessionServer: sesssrv.NewConfig(),
		SOMC:          somc.NewConfig(),
		UI:            ui.NewConfig(),
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

func createLogger(conf *config, db *reform.DB) (log2.Logger, error) {
	flog, err := log2.NewStderrLogger(conf.FileLog)
	if err != nil {
		return nil, err
	}

	dlog, err := dblog.NewLogger(conf.DBLog, db)
	if err != nil {
		return nil, err
	}

	return log2.NewMultiLogger(flog, dlog), nil
}

func createUIServer(conf *ui.Config, logger log2.Logger,
	db *reform.DB, queue job.Queue) (*rpcsrv.Server, error) {
	server, err := rpcsrv.NewServer(conf.Config)
	if err != nil {
		return nil, err
	}

	handler := ui.NewHandler(conf, logger, db, nil)
	if err := server.AddHandler("ui", handler); err != nil {
		return nil, err
	}

	return server, nil
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

	db, err := data.NewDB(conf.DB)
	if err != nil {
		logger.Fatal("failed to open db connection: %s", err)
	}
	defer data.CloseDB(db)

	logger2, err := createLogger(conf, db)
	if err != nil {
		logger.Fatal("failed to create logger: %s", err)
	}

	reporter, err := bugsnag.NewClient(conf.Report, db, logger)
	if err != nil {
		logger.Fatal("failed to create Bugsnag client: %s", err)
	}

	logger.Reporter(reporter)

	ethClient, err := eth.NewClient(context.Background(), conf.Eth)
	if err != nil {
		logger.Fatal("failed to dial Ethereum node: %v", err)
	}
	defer ethClient.Close()

	ptcAddr := common.HexToAddress(conf.Eth.Contract.PTCAddrHex)
	ptc, err := contract.NewPrivatixTokenContract(
		ptcAddr, ethClient.EthClient())
	if err != nil {
		logger.Fatal("failed to create ptc instance: %v", err)
	}

	pscAddr := common.HexToAddress(conf.Eth.Contract.PSCAddrHex)
	psc, err := contract.NewPrivatixServiceContract(
		pscAddr, ethClient.EthClient())
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
		worker.NewEthBackend(psc, ptc, ethClient.EthClient(),
			conf.Eth.Timeout), conf.Gas,
		pscAddr, conf.PayAddress, pwdStorage, data.ToPrivateKey,
		conf.EptMsg)
	if err != nil {
		logger.Fatal("failed to create worker: %s", err)
	}

	queue := job.NewQueue(conf.Job, logger, db, handlers.HandlersMap(worker))
	defer queue.Close()
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, db, queue)
	worker.SetProcessor(pr)

	runner := svcrun.NewServiceRunner(conf.ServiceRunner, logger, db, pr)
	defer runner.StopAll()
	worker.SetRunner(runner)

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, queue, pwdStorage, pr)
	go func() {
		fatal <- fmt.Sprintf("failed to run agent server: %s\n",
			uiSrv.ListenAndServe())
	}()

	uiSrv2, err := createUIServer(conf.UI, logger2, db, queue)
	if err != nil {
		logger.Fatal("failed to create UI server: %s", err)
	}
	go func() {
		fatal <- fmt.Sprintf("failed to run UI server: %s\n",
			uiSrv2.ListenAndServe())
	}()

	amon, err := abill.NewMonitor(conf.AgentMonitor.Interval,
		db, logger, pr)
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

	mon, err := monitor.NewMonitor(conf.BlockMonitor, logger2, db, queue,
		conf.Eth, pscAddr, ptcAddr, ethClient.EthClient(),
		ethClient.CloseIdleConnections)
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
