package report

import (
	"os"
	"path/filepath"

	"github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
	"gopkg.in/reform.v1"
)

const (
	currentAPIKey = "c021f92e9c199c79d870adf34365e372"
	currentStage  = AlphaStage
	mainRepo      = "github.com/privatix/dappctrl"
)

// Stages of application
const (
	AlphaStage = iota
	BettaStage
	RCStage
	RTMStage
)

var (
	notifier *bugsnag.Notifier

	// Enable if this is true then send errors to a remote service
	Enable bool

	db *reform.DB

	lastAcc string

	logger Log

	// This slice is needed so that the full path is written to the log
	pkgSlice = []string{"main", "agent/billing", "client/bill", "data",
		"eth", "eth/contract", "eth/truffle", "eth/util",
		"execsrv", "job", "messages", "messages/ept",
		"messages/ept/config", "messages/offer",
		"monitor", "pay", "proc", "worker", "sesssrv", "somc",
		"svc/dappvpn", "svc/mon", "svc/pusher", "uisrv", "util",
		"util/srv"}

	defaultAppID  = emptyUUID()
	defaultAccEth = new(common.Address).String()
)

// Config Bugsnag client config
type Config struct {
	AppID string
}

// Log interface for report
type Log interface {
	Printf(format string, v ...interface{})
	Debug(fmt string, v ...interface{})
}

// NewConfig generates a new default Bugsnag client Config.
func NewConfig() *Config {
	return &Config{AppID: defaultAppID}
}

// it is here because of cross-import with utils
func isUUID(s string) bool {
	_, err := uuid.FromString(s)
	return err == nil
}

func emptyUUID() string {
	return new(uuid.UUID).String()
}

func genAcc() string {
	if lastAcc == "" {
		return defaultAccEth
	}
	return lastAcc
}

func metadata(update bool) bugsnag.MetaData {
	return bugsnag.MetaData{
		"Account": {
			"EthAddr": genAcc(),
		},
	}
}

func user(appID string) bugsnag.User {
	return bugsnag.User{Id: app(appID)}
}

func app(appID string) string {
	if appID == "" || isUUID(appID) {
		return defaultAppID
	}
	return appID
}

func stageToStr(stage int) string {
	var result string

	switch stage {
	case AlphaStage:
		result = "alpha"
	case BettaStage:
		result = "betta"
	default:
		result = "alpha"
	}
	return result
}

/*func Acc() string {
}*/

// NewReporter initializing Bugsnag client
func NewReporter(cfg *Config, database *reform.DB, log Log) {
	if database == nil || log == nil {
		return
	}

	for k, v := range pkgSlice {
		// if you do not add an *,
		// the full path will not be displayed in the dashboard
		pkgSlice[k] = filepath.Join(mainRepo, v) + "*"
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:          currentAPIKey,
		ReleaseStage:    stageToStr(currentStage),
		ProjectPackages: pkgSlice,
		Logger:          log,
		PanicHandler:    func() {}, // we use our panic processor
	})

	notifier = bugsnag.New(user(cfg.AppID))

	db = database
	logger = log

	Enable = true
}

// Notify takes three arguments: err - standard error;
// sync - if true then the function waits for the end of sending;
// skip - how many errors to remove from stacktrace.
func Notify(err error, sync bool, skip int) {
	if Enable {
		notify(err, sync, skip)
	}
}

func notify(err error, sync bool, skip int) {
	var e error
	if sync {
		e = notifier.NotifySync(errors.New(err, skip), true, metadata(true))
	} else {
		e = notifier.Notify(errors.New(err, skip), metadata(true))
	}
	if e != nil {
		logger.Debug("failed to send notify: %s", err)
	}
}

// PanicHunter catches panic, in case of an enabled reporter
func PanicHunter() {
	if err := recover(); err != nil {
		if Enable {
			notifier.NotifySync(errors.New(err, 3), true, metadata(false))
			os.Exit(1)
		}
		panic(err)
	}
}
