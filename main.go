package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"gopkg.in/reform.v1"

	abill "github.com/privatix/dappctrl/agent/bill"
	"github.com/privatix/dappctrl/agent/somcsrv"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/somc"
	"github.com/privatix/dappctrl/country"
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
	"github.com/privatix/dappctrl/proc/looper"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	rlog "github.com/privatix/dappctrl/report/log"
	"github.com/privatix/dappctrl/sesssrv"
	"github.com/privatix/dappctrl/ui"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
	"github.com/privatix/dappctrl/version"
)

// Values for versioning.
var (
	Commit  string
	Version string
)

type config struct {
	AgentMonitor     *abill.Config
	BlockMonitor     *monitor.Config
	ClientMonitor    *cbill.Config
	Country          *country.Config
	DB               *data.DBConfig
	DBLog            *dblog.Config
	ReportLog        *rlog.Config
	EptMsg           *ept.Config
	Eth              *eth.Config
	FileLog          *log.FileConfig
	Gas              *worker.GasConf
	Job              *job.Config
	Looper           *looper.Config
	PayServer        *pay.Config
	PayAddress       string
	Proc             *proc.Config
	Report           *bugsnag.Config
	Role             string
	SessionServer    *sesssrv.Config
	SOMCServer       *rpcsrv.Config
	StaticPassword   string
	TorHostname      string
	TorSocksListener uint
	UI               *rpcsrv.Config
}

func newConfig() *config {
	return &config{
		AgentMonitor:  abill.NewConfig(),
		BlockMonitor:  monitor.NewConfig(),
		ClientMonitor: cbill.NewConfig(),
		Country:       country.NewConfig(),
		DB:            data.NewDBConfig(),
		DBLog:         dblog.NewConfig(),
		Looper:        looper.NewConfig(),
		ReportLog:     rlog.NewConfig(),
		EptMsg:        ept.NewConfig(),
		Eth:           eth.NewConfig(),
		FileLog:       log.NewFileConfig(),
		Job:           job.NewConfig(),
		PayServer:     pay.NewConfig(),
		Proc:          proc.NewConfig(),
		Report:        bugsnag.NewConfig(),
		SessionServer: sesssrv.NewConfig(),
		SOMCServer:    rpcsrv.NewConfig(),
		UI:            rpcsrv.NewConfig(),
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
	if conf.StaticPassword == "" {
		return new(data.PWDStorage)
	}
	storage := data.StaticPWDStorage(conf.StaticPassword)
	return &storage
}

func createLogger(conf *config, db *reform.DB) (log.Logger, io.Closer, error) {
	elog, err := log.NewStderrLogger(conf.FileLog.WriterConfig)
	if err != nil {
		return nil, nil, err
	}

	flog, closer, err := log.NewFileLogger(conf.FileLog)
	if err != nil {
		return nil, nil, err
	}

	dlog, err := dblog.NewLogger(conf.DBLog, db)
	if err != nil {
		return nil, nil, err
	}

	blog, err := rlog.NewLogger(conf.ReportLog)
	if err != nil {
		return nil, nil, err
	}

	logger := log.NewMultiLogger(flog, elog, dlog, blog)

	blog2 := blog.(bugsnag.Log)
	reporter, err := bugsnag.NewClient(conf.Report, db, blog2,
		version.Message(Commit, Version))
	if err != nil {
		return nil, nil, err
	}
	blog2.Reporter(reporter)
	blog2.Logger(logger)

	return logger, closer, nil
}

func createUIServer(conf *rpcsrv.Config, logger log.Logger, db *reform.DB,
	queue job.Queue, pwdStorage data.PWDGetSetter, userRole string,
	processor *proc.Processor, somcClientBuilder somc.ClientBuilderInterface) (*rpcsrv.Server, error) {
	server, err := rpcsrv.NewServer(conf)
	if err != nil {
		return nil, err
	}

	handler := ui.NewHandler(logger, db, queue, pwdStorage,
		data.EncryptedKey, data.ToPrivateKey, userRole, processor,
		somcClientBuilder)
	if err := server.AddHandler("ui", handler); err != nil {
		return nil, err
	}

	return server, nil
}

func startAutoPopUpLoop(ctx context.Context, cfg *looper.Config,
	logger log.Logger, db *reform.DB, queue job.Queue,
	ethBack eth.Backend) error {
	period, err := data.ReadUintSetting(db.Querier, data.SettingsPeriodPopUp)
	if err != nil {
		return err
	}

	var timeout time.Duration
	if cfg.AutoOfferingPopUpTimeout != 0 {
		timeout = time.Millisecond *
			time.Duration(cfg.AutoOfferingPopUpTimeout)
	} else {
		timeout = time.Duration(period) * looper.BlockTime / 2

	}

	serviceContractABI, err := abi.JSON(strings.NewReader(
		contract.PrivatixServiceContractABI))
	if err != nil {
		return err
	}

	autoPopUpOfferingFunc := func() []*data.Job {
		return looper.AutoOfferingPopUp(logger, serviceContractABI,
			db, ethBack, time.Now, period)
	}

	looper.Loop(ctx, logger, db, queue, timeout, autoPopUpOfferingFunc)

	return err
}

func panicHunter(logger log.Logger) {
	if err := recover(); err != nil {
		logger.Fatal(fmt.Sprintf("panic raised: %+v", err))
	}
}

func newAgentSOMCServer(conf *rpcsrv.Config, db *reform.DB,
	logger log.Logger) (*rpcsrv.Server, error) {
	server, err := rpcsrv.NewServer(conf)
	if err != nil {
		return nil, err
	}

	handler := somcsrv.NewHandler(db, logger)
	if err := server.AddHandler("api", handler); err != nil {
		return nil, err
	}

	return server, nil
}

func main() {
	if err := data.ExecuteCommand(os.Args[1:]); err != nil {
		panic(fmt.Sprintf("failed to execute command: %s", err))
	}

	fatal := make(chan error)
	defer bugsnag.PanicHunter()

	conf := newConfig()
	readFlags(conf)

	db, err := data.NewDB(conf.DB)
	if err != nil {
		panic(fmt.Sprintf("failed to open db "+
			"connection: %s", err))
	}
	defer data.CloseDB(db)

	logger, closer, err := createLogger(conf, db)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %s", err))
	}
	defer closer.Close()
	defer panicHunter(logger)

	if err := data.Recover(db); err != nil {
		logger.Fatal(err.Error())
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill)

	go func() {
		<-interrupt
		// Delay to execute deferred operations.
		time.Sleep(time.Second * 3)
		logger.Debug("dappctrl is stopped")
		os.Exit(1)
	}()

	pwdStorage := getPWDStorage(conf)

	ethBack := eth.NewBackend(conf.Eth, logger)

	worker, err := worker.NewWorker(logger, db, ethBack, conf.Gas,
		ethBack.PSCAddress(), conf.PayAddress, pwdStorage, conf.Country,
		data.ToPrivateKey, conf.EptMsg, conf.TorHostname,
		somc.NewClientBuilder(conf.TorSocksListener))
	if err != nil {
		logger.Fatal(err.Error())
	}

	queue := job.NewQueue(conf.Job, logger, db, handlers.HandlersMap(worker))
	defer queue.Close()
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, db, queue)
	worker.SetProcessor(pr)

	mon, err := monitor.NewMonitor(conf.BlockMonitor, ethBack, db, logger,
		ethBack.PSCAddress(), ethBack.PTCAddress(), conf.Role, queue)
	if err != nil {
		logger.Fatal(err.Error())
	}
	mon.Start()
	defer mon.Stop()

	go func() {
		fatal <- queue.Process()
	}()

	uiSrv, err := createUIServer(conf.UI, logger, db, queue, pwdStorage,
		conf.Role, pr, somc.NewClientBuilder(conf.TorSocksListener))
	if err != nil {
		logger.Fatal(err.Error())
	}
	go func() {
		fatal <- uiSrv.ListenAndServe()
	}()

	sess := sesssrv.NewServer(
		conf.SessionServer, logger, db, conf.Country, queue)
	go func() {
		fatal <- sess.ListenAndServe()
	}()
	defer sess.Close()

	if conf.Role == data.RoleClient {
		cmon := cbill.NewMonitor(conf.ClientMonitor, logger, db, pr,
			conf.Eth.Contract.PSCAddrHex, pwdStorage)
		go func() {
			fatal <- cmon.Run()
		}()
		defer cmon.Close()
	}

	if conf.Role == data.RoleAgent {
		somcServer, err := newAgentSOMCServer(conf.SOMCServer, db, logger)
		if err != nil {
			logger.Fatal(err.Error())
		}
		go func() {
			fatal <- somcServer.ListenAndServe()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = startAutoPopUpLoop(ctx, conf.Looper,
			logger, db, queue, ethBack)
		if err != nil {
			logger.Fatal(err.Error())
		}

		paySrv := pay.NewServer(conf.PayServer, logger, db)
		go func() {
			fatal <- paySrv.ListenAndServe()
		}()
		defer paySrv.Close()

		amon, err := abill.NewMonitor(conf.AgentMonitor.Interval,
			db, logger, pr)
		if err != nil {
			logger.Fatal(err.Error())
		}
		go func() {
			fatal <- amon.Run()
		}()
	}

	err = <-fatal
	logger.Fatal(err.Error())
}
