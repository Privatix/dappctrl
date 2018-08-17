package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/svcrun"
	"github.com/privatix/dappctrl/data"
	dblog "github.com/privatix/dappctrl/data/log"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/internal/version"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/handlers"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	rlog "github.com/privatix/dappctrl/report/log"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/uisrv"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

type config struct {
	AgentMonitor  *abill.Config
	AgentServer   *uisrv.Config
	BlockMonitor  *monitor.Config
	ClientMonitor *cbill.Config
	DB            *data.DBConfig
	DBLog         *dblog.Config
	ReportLog     *rlog.Config
	EptMsg        *ept.Config
	Eth           *eth.Config
	FileLog       *log.FileConfig
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
		ReportLog:     rlog.NewConfig(),
		EptMsg:        ept.NewConfig(),
		Eth:           eth.NewConfig(),
		FileLog:       log.NewFileConfig(),
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

func readFlags(conf *config) {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	v := flag.Bool("version", false, "Prints current dappctrl version")

	flag.Parse()

	version.Print(*v, Commit, Version)

	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s", err))
	}
}

func getPWDStorage(conf *config) data.PWDGetSetter {
	if conf.StaticPasword == "" {
		return new(data.PWDStorage)
	}
	storage := data.StaticPWDStorage(conf.StaticPasword)
	return &storage
}

func createLogger(conf *config, db *reform.DB) (log.Logger, bugsnag.Log, error) {
	flog, err := log.NewStderrLogger(conf.FileLog)
	if err != nil {
		return nil, nil, err
	}

	dlog, err := dblog.NewLogger(conf.DBLog, db)
	if err != nil {
		return nil, nil, err
	}

	rLog, err := rlog.NewLogger(conf.ReportLog)
	if err != nil {
		return nil, nil, err
	}

	return log.NewMultiLogger(rLog, flog, dlog), rLog.(bugsnag.Log), nil
}

func createUIServer(conf *ui.Config, logger log.Logger,
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
	fatal := make(chan error)
	defer bugsnag.PanicHunter()

	conf := newConfig()
	readFlags(conf)

	logger, err := util.NewLogger(conf.Log)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}
	defer logger.GracefulStop()

	db, err := data.NewDB(conf.DB)
	if err != nil {
		panic(fmt.Sprintf("failed to open db "+
			"connection: %s", err))
	}
	defer data.CloseDB(db)

	logger2, rLog, err := createLogger(conf, db)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}

	reporter, err := bugsnag.NewClient(conf.Report, db, rLog)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	rLog.Reporter(reporter)
	rLog.Logger(logger2)

	ethClient, err := eth.NewClient(context.Background(),
		conf.Eth, logger2)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	defer ethClient.Close()

	ptcAddr := common.HexToAddress(conf.Eth.Contract.PTCAddrHex)
	ptc, err := contract.NewPrivatixTokenContract(
		ptcAddr, ethClient.EthClient())
	if err != nil {
		logger2.Fatal(err.Error())
	}

	pscAddr := common.HexToAddress(conf.Eth.Contract.PSCAddrHex)
	psc, err := contract.NewPrivatixServiceContract(
		pscAddr, ethClient.EthClient())
	if err != nil {
		logger2.Fatal(err.Error())
	}

	paySrv := pay.NewServer(conf.PayServer, logger2, db)
	go func() {
		fatal <- paySrv.ListenAndServe()
	}()
	defer paySrv.Close()

	sess := sesssrv.NewServer(conf.SessionServer, logger2, db)
	go func() {
		fatal <- sess.ListenAndServe()
	}()
	defer sess.Close()

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	defer somcConn.Close()

	pwdStorage := getPWDStorage(conf)

	worker, err := worker.NewWorker(logger2, db, somcConn, 
		worker.NewEthBackend(psc, ptc, ethClient.EthClient(),
			conf.Eth.Timeout), conf.Gas,
		pscAddr, conf.PayAddress, pwdStorage, data.ToPrivateKey,
		conf.EptMsg)
	if err != nil {
		logger2.Fatal(err.Error())
	}

	queue := job.NewQueue(conf.Job, logger2, db, handlers.HandlersMap(worker))
	defer queue.Close()
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, db, queue)
	worker.SetProcessor(pr)

	runner := svcrun.NewServiceRunner(conf.ServiceRunner, logger, db, pr)
	defer runner.StopAll()
	worker.SetRunner(runner)

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, queue, pwdStorage, pr)
	go func() {
		fatal <- uiSrv.ListenAndServe()
	}()

	uiSrv2, err := createUIServer(conf.UI, logger2, db, queue)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	go func() {
		fatal <- uiSrv2.ListenAndServe()
	}()

	amon, err := abill.NewMonitor(conf.AgentMonitor.Interval,
		db, logger2, pr)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	go func() {
		fatal <- amon.Run()
	}()

	cmon := cbill.NewMonitor(conf.ClientMonitor,
		logger2, db, pr, conf.Eth.Contract.PSCAddrHex, pwdStorage)
	go func() {
		fatal <- cmon.Run()
	}()
	defer cmon.Close()

	mon, err := monitor.NewMonitor(conf.BlockMonitor, logger2, db, queue,
		conf.Eth, pscAddr, ptcAddr, ethClient.EthClient(),
		ethClient.CloseIdleConnections)
	if err != nil {
		logger2.Fatal(err.Error())
	}
	if err := mon.Start(); err != nil {
		logger2.Fatal(err.Error())
	}
	defer mon.Stop()

	go func() {
		fatal <- queue.Process()
	}()

	err = <-fatal
	logger2.Fatal(err.Error())
}
