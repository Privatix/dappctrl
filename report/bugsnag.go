package report

import (
	"path/filepath"

	"github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
	"gopkg.in/reform.v1"
)

const (
	apiKey       = "c021f92e9c199c79d870adf34365e372"
	currentStage = AlphaStage
	mainRepo     = "github.com/privatix/dappctrl"
)

const (
	AlphaStage = iota
	BettaStage
)

var (
	// Config global configuration bugsnag client
	Config bugsnag.Configuration

	// Enable if this is true then send errors to a remote service
	Enable bool

	appID string

	db *reform.DB

	configured bool

	logger Log

	// This slice is needed so that the full path is written to the log
	pkgSlice = []string{"main", "agent/billing", "client/bill", "data",
		"eth", "eth/contract", "eth/truffle", "eth/util",
		"execsrv", "job", "messages", "messages/ept",
		"messages/ept/config", "messages/offer",
		"monitor", "pay", "proc", "worker", "sesssrv", "somc",
		"svc/dappvpn", "svc/mon", "svc/pusher", "uisrv", "util",
		"util/srv"}

	RawData interface{}
)

// Log interface for report
type Log interface {
	Printf(format string, v ...interface{})
	Debug(fmt string, v ...interface{})
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

/*func defaultAcc() string {
}*/

// NewReporter
func NewReporter(applicationID string, database *reform.DB, log Log) {
	if applicationID == "" || database == nil || log == nil {
		return
	}

	for k, v := range pkgSlice {
		// if you do not add an *,
		// the full path will not be displayed in the dashboard
		pkgSlice[k] = filepath.Join(mainRepo, v) + "*"
	}

	Config = bugsnag.Configuration{
		APIKey:          apiKey,
		ReleaseStage:    stageToStr(currentStage),
		ProjectPackages: pkgSlice,
		Logger:          log,
		Synchronous:     true,
	}
	bugsnag.Configure(Config)

	db = database
	logger = log

	appID = applicationID

	RawData = []interface{}{
		bugsnag.User{Id: appID},
		bugsnag.MetaData{
			"Account": {
				"EthAddr": "0x123", // TODO: [maxim] this should be taken from the database
			},
		},
	}

	configured = true
}

func Notify(err error) {
	if configured {
		notify(err)
	}
}

func notify(e error) {
	// modify stacktrace, skip: bugsnag.notify, bugsnag.Notify,
	// log.Log, log.Warn or log.Error or log.Fatal.
	if err := bugsnag.Notify(errors.New(e, 4),
		bugsnag.User{Id: appID}, bugsnag.MetaData{
			"Account": {
				"EthAddr": "0x123", // TODO: [maxim] this should be taken from the database
			},
		}); err != nil {
		logger.Debug("failed to send notify: %s", err)
	}
}

func PanicHunter() {

}
