package assemble

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
	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	dblog "github.com/privatix/dappctrl/data/log"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/eth/contract"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/handlers"
	"github.com/privatix/dappctrl/proc/looper"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	rlog "github.com/privatix/dappctrl/report/log"
	"github.com/privatix/dappctrl/sess"
	"github.com/privatix/dappctrl/somc"
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

func readFlags(conf *Config) {
	fconfig := flag.String(
		"config", "dappctrl.config.json", "Configuration file")
	v := flag.Bool("version", false, "Prints current dappctrl version")

	flag.Parse()

	version.Print(*v, Commit, Version)

	if err := util.ReadJSONFile(*fconfig, &conf); err != nil {
		panic(fmt.Sprintf("failed to read configuration: %s", err))
	}
}

func getPWDStorage(conf *Config) data.PWDGetSetter {
	if conf.StaticPassword == "" {
		return new(data.PWDStorage)
	}
	storage := data.StaticPWDStorage(conf.StaticPassword)
	return &storage
}

func createLogger(conf *Config, db *reform.DB) (log.Logger, io.Closer, error) {
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
		somcClientBuilder, ui.NewSimpleToken())
	if err := server.AddHandler("ui", handler); err != nil {
		return nil, err
	}

	return server, nil
}

func createSessServer(conf *rpcsrv.Config, logger log.Logger, db *reform.DB,
	countryConf *country.Config, queue job.Queue) (*rpcsrv.Server, error) {
	server, err := rpcsrv.NewServer(conf)
	if err != nil {
		return nil, err
	}

	handler := sess.NewHandler(logger, db, countryConf, queue)
	if err := server.AddHandler("sess", handler); err != nil {
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

// RunApp runs application.
func RunApp(fatal chan error) {
	conf := NewConfig()
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

	somcProps := somc.NewProps(conf.AgentTorSOMC, conf.AgentDirectSOMC, db)

	worker, err := worker.NewWorker(logger, db, ethBack, conf.Gas,
		ethBack.PSCAddress(), conf.PayAddress, pwdStorage, conf.Country,
		data.ToPrivateKey, conf.EptMsg, somcProps,
		somc.NewClientBuilder(conf.ClientTorSOMC, db))
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
		conf.Role, pr, somc.NewClientBuilder(conf.ClientTorSOMC, db))
	if err != nil {
		logger.Fatal(err.Error())
	}
	go func() {
		fatal <- uiSrv.ListenAndServe()
	}()

	sessSrv, err := createSessServer(
		conf.Sess, logger, db, conf.Country, queue)
	if err != nil {
		logger.Fatal(err.Error())
	}
	go func() {
		fatal <- sessSrv.ListenAndServe()
	}()

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
