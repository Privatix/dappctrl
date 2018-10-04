package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/reform.v1"

	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/client/svcrun"
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
	"github.com/privatix/dappctrl/proc/adapter"
	"github.com/privatix/dappctrl/proc/handlers"
	"github.com/privatix/dappctrl/proc/looper"
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
	"github.com/privatix/dappctrl/version"
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
	PayServer     *pay.Config
	PayAddress    string
	Proc          *proc.Config
	Report        *bugsnag.Config
	Role          string
	ServiceRunner *svcrun.Config
	SessionServer *sesssrv.Config
	SOMC          *somc.Config
	StaticPasword string
	UI            *ui.Config
	Country       *country.Config
	Looper        *looper.Config
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
		Proc:          proc.NewConfig(),
		Report:        bugsnag.NewConfig(),
		ServiceRunner: svcrun.NewConfig(),
		SessionServer: sesssrv.NewConfig(),
		SOMC:          somc.NewConfig(),
		UI:            ui.NewConfig(),
		Country:       country.NewConfig(),
		Looper:        looper.NewConfig(),
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

	logger := log.NewMultiLogger(elog, flog, dlog, blog)

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

func createUIServer(conf *ui.Config, logger log.Logger, db *reform.DB,
	queue job.Queue, pwdStorage data.PWDGetSetter) (*rpcsrv.Server, error) {
	server, err := rpcsrv.NewServer(conf.Config)
	if err != nil {
		return nil, err
	}

	handler := ui.NewHandler(conf, logger, db, queue, pwdStorage,
		data.EncryptedKey, data.ToPrivateKey)
	if err := server.AddHandler("ui", handler); err != nil {
		return nil, err
	}

	return server, nil
}

func startAutoPopUpLoop(ctx context.Context, cfg *looper.Config,
	logger log.Logger, db *reform.DB, queue job.Queue,
	ethBack adapter.EthBackend) error {
	popUpPeriod, err := ethBack.PSCGetPopUpPeriod(&bind.CallOpts{})
	if err != nil {
		return err
	}

	var timeout time.Duration
	if cfg.AutoOfferingPopUpTimeout != 0 {
		timeout = time.Second *
			time.Duration(cfg.AutoOfferingPopUpTimeout)
	} else {
		timeout = time.Duration(popUpPeriod) * looper.BlockTime / 2

	}

	serviceContractABI, err := abi.JSON(strings.NewReader(
		contract.PrivatixServiceContractABI))
	if err != nil {
		return err
	}

	autoPopUpOfferingFunc := func() []*data.Job {
		return looper.AutoOfferingPopUp(
			logger, serviceContractABI, db, ethBack, time.Now)
	}

	looper.Loop(ctx, logger, db, queue, timeout, autoPopUpOfferingFunc)

	return err
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

	ethClient, err := eth.NewClient(context.Background(),
		conf.Eth, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer ethClient.Close()

	ptcAddr := common.HexToAddress(conf.Eth.Contract.PTCAddrHex)
	ptc, err := contract.NewPrivatixTokenContract(
		ptcAddr, ethClient.EthClient())
	if err != nil {
		logger.Fatal(err.Error())
	}

	pscAddr := common.HexToAddress(conf.Eth.Contract.PSCAddrHex)
	psc, err := contract.NewPrivatixServiceContract(
		pscAddr, ethClient.EthClient())
	if err != nil {
		logger.Fatal(err.Error())
	}

	somcConn, err := somc.NewConn(conf.SOMC, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer somcConn.Close()

	pwdStorage := getPWDStorage(conf)

	ethBack := adapter.NewEthBackend(
		psc, ptc, ethClient.EthClient(), conf.Eth.Timeout)

	worker, err := worker.NewWorker(logger, db, somcConn,
		ethBack, conf.Gas, pscAddr, conf.PayAddress,
		pwdStorage, conf.Country, data.ToPrivateKey, conf.EptMsg)
	if err != nil {
		logger.Fatal(err.Error())
	}

	queue := job.NewQueue(conf.Job, logger, db, handlers.HandlersMap(worker))
	defer queue.Close()
	worker.SetQueue(queue)

	pr := proc.NewProcessor(conf.Proc, db, queue)
	worker.SetProcessor(pr)

	runner := svcrun.NewServiceRunner(conf.ServiceRunner, logger, db, pr)
	defer runner.StopAll()
	worker.SetRunner(runner)

	mon, err := monitor.NewMonitor(conf.BlockMonitor, logger, db, queue,
		conf.Eth, pscAddr, ptcAddr, ethClient.EthClient(),
		conf.Role, ethClient.CloseIdleConnections)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if err := mon.Start(); err != nil {
		logger.Fatal(err.Error())
	}
	defer mon.Stop()

	go func() {
		fatal <- queue.Process()
	}()

	uiSrv := uisrv.NewServer(conf.AgentServer, logger, db, conf.Role,
		queue, pwdStorage, pr)
	go func() {
		fatal <- uiSrv.ListenAndServe()
	}()

	uiSrv2, err := createUIServer(conf.UI, logger, db, queue, pwdStorage)
	if err != nil {
		logger.Fatal(err.Error())
	}
	go func() {
		fatal <- uiSrv2.ListenAndServe()
	}()

	if conf.Role == data.RoleClient {
		cmon := cbill.NewMonitor(conf.ClientMonitor, logger, db, pr,
			conf.Eth.Contract.PSCAddrHex, pwdStorage)
		go func() {
			fatal <- cmon.Run()
		}()
		defer cmon.Close()
	}

	sess := sesssrv.NewServer(conf.SessionServer, logger, db, conf.Country)
	go func() {
		fatal <- sess.ListenAndServe()
	}()
	defer sess.Close()

	if conf.Role == data.RoleAgent {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = startAutoPopUpLoop(
			ctx, conf.Looper, logger, db, queue, ethBack)
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
