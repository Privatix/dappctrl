package bugsnag

import (
	"database/sql"
	"path/filepath"
	"strconv"

	"github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
)

const (
	account       = "Account"
	ethAddress    = "EthAddr"
	currentAPIKey = "c021f92e9c199c79d870adf34365e372"
	currentStage  = alphaStage
	mainRepo      = "github.com/privatix/dappctrl"

	alpha = "alpha"
	betta = "betta"
	rc    = "rc"
	rtm   = "rtm"
)

// Stages of application.
const (
	alphaStage = iota
	bettaStage
	rcStage
	rtmStage
)

const (
	key = "error.sendremote"
)

var (
	defaultAccEth = new(common.Address).String()
	defaultAppID  = emptyUUID()

	enable bool

	notifier *bugsnag.Notifier

	// TODO(maxim) The list needs to be configured dynamically, before the application starts
	// This slice is needed so that the full path is written to the log
	pkgSlice = []string{"main", "agent/billing", "client/bill", "data",
		"eth", "eth/contract", "eth/truffle", "eth/util",
		"execsrv", "job", "messages", "messages/ept",
		"messages/ept/config", "messages/offer",
		"monitor", "pay", "proc", "worker", "sesssrv", "somc",
		"svc/dappvpn", "svc/mon", "svc/pusher", "uisrv", "util",
		"util/srv"}
)

// Log interface for report.
type Log interface {
	Debug(fmt string, v ...interface{})
	Printf(format string, v ...interface{})
	Warn(fmt string, v ...interface{})
}

// Config Bugsnag client config.
type Config struct {
	AppID string
}

// Client Bugsnag client object.
type Client struct {
	db       *reform.DB
	enable   bool
	lastAcc  string
	logger   Log
	notifier *bugsnag.Notifier
}

// NewConfig generates a new default Bugsnag client Config.
func NewConfig() *Config {
	return &Config{AppID: defaultAppID}
}

// NewClient initializing Bugsnag client.
func NewClient(cfg *Config, db *reform.DB, log Log) *Client {
	if log == nil {
		return nil
	}

	for k, v := range pkgSlice {
		// if you do not add an *,
		// the full path will not be displayed in the dashboard
		pkgSlice[k] = filepath.Join(mainRepo, v) + "*"
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:          currentAPIKey,
		Logger:          log,
		PanicHandler:    func() {}, // we use our panic processor
		ProjectPackages: pkgSlice,
		ReleaseStage:    stageToStr(currentStage),
	})

	cli := new(Client)
	cli.db = db
	cli.logger = log
	cli.notifier = bugsnag.New(user(cfg.AppID))

	//check enable service
	e := cli.allowed()
	cli.enable = e
	enable = e
	notifier = cli.notifier
	return cli
}

func emptyUUID() string {
	return new(uuid.UUID).String()
}

func metadata(ethAddr string) bugsnag.MetaData {
	return bugsnag.MetaData{
		account: {
			ethAddress: ethAddr,
		},
	}
}

func accEthAddr(db *reform.DB) string {
	var tempAddr string
	if err := db.QueryRow(`
		           SELECT eth_addr
                             FROM accounts
                            ORDER BY is_default
                            LIMIT 1;`).Scan(&tempAddr); err != nil {
		return defaultAccEth
	}
	addr, err := data.ToAddress(tempAddr)
	if err != nil {
		return defaultAccEth
	}
	return addr.String()
}

func user(appID string) bugsnag.User {
	return bugsnag.User{Id: app(appID)}
}

func app(appID string) string {
	if appID == "" || !util.IsUUID(appID) {
		return defaultAppID
	}
	return appID
}

func stageToStr(stage int) string {
	var result string

	switch stage {
	case alphaStage:
		result = alpha
	case bettaStage:
		result = betta
	case rcStage:
		result = rc
	case rtmStage:
		result = rtm
	default:
		result = rtm
	}
	return result
}

// Notify takes three arguments:
// err - standard error;
// sync - if true then the function waits for the end of sending;
// skip - how many errors to remove from stacktrace.
func (c *Client) Notify(err error, sync bool, skip int) {
	if c.enable {
		c.notify(err, sync, skip)
	}
}

func (c *Client) notify(err error, sync bool, skip int) {
	ethAdd := accEthAddr(c.db)

	var e error
	if sync {
		e = c.notifier.NotifySync(errors.New(err, skip),
			true, metadata(ethAdd))
	} else {
		e = c.notifier.Notify(errors.New(err, skip),
			metadata(ethAdd))
	}
	if e != nil {
		c.logger.Debug("failed to send notify: %s", e)
	}
}

// Enable client is enabled?
func (c *Client) Enable() bool {
	return c.enable
}

func (c *Client) allowed() bool {
	var setting data.Setting
	err := c.db.FindByPrimaryKeyTo(&setting, key)
	if err != nil {
		if err == sql.ErrNoRows {
			c.logger.Warn("key %s is not exist"+
				" in Setting table", key)
		} else {
			c.logger.Warn("failed to get key %s"+
				" from Setting table", key)
		}
		return false
	}
	val, err := strconv.ParseBool(setting.Value)
	if err != nil {
		c.logger.Warn("key %s from Setting table"+
			" has an incorrect format", key)
		return false
	}
	return val
}

// PanicHunter catches panic, in case of an enabled reporter.
func PanicHunter() {
	if err := recover(); err != nil {
		if enable && notifier != nil {
			notifier.NotifySync(
				errors.New(err, 3), true,
				metadata(defaultAccEth))
		}
		panic(err)
	}
}
