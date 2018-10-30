package bugsnag

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/bugsnag/bugsnag-go"
	"github.com/bugsnag/bugsnag-go/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/report"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

const (
	accounts            = "Accounts"
	defaultReleaseStage = "development"
	production          = "production"
	staging             = "staging"

	currentAPIKey = "5d98bc82cbbd99bcd413cb67dbe823bb"
	key           = "error.sendremote"
)

var (
	defaultAccEth = new(common.Address).String()
	defaultAppID  = emptyUUID()

	enable      bool
	panicIgnore bool

	notifier *bugsnag.Notifier
)

// Log interface for report.
type Log interface {
	log.Logger
	Printf(format string, v ...interface{})
	Reporter(reporter report.Reporter)
	Logger(logger log.Logger)
}

// Config Bugsnag client config.
type Config struct {
	AppID            string
	ReleaseStage     string
	ExcludedPackages []string
}

// Client Bugsnag client object.
type Client struct {
	db       *reform.DB
	enable   bool
	logger   Log
	notifier *bugsnag.Notifier
}

// NewConfig generates a new default Bugsnag client Config.
func NewConfig() *Config {
	return &Config{AppID: defaultAppID, ReleaseStage: defaultReleaseStage}
}

// NewClient initializing Bugsnag client.
// Bugsnag client that automatic error sender to remote server.
// We use this service to collect anonymous information
// about the error and panic.
// Service is activated if exist entry key = "error.sendremote"
// and value = true in the database settings table.
func NewClient(
	cfg *Config, db *reform.DB, log Log, version string) (*Client, error) {
	if log == nil {
		return nil, fmt.Errorf("no log object specified")
	}

	excludedPackagesMap := make(map[string]bool)

	for _, pkg := range cfg.ExcludedPackages {
		excludedPackagesMap[pkg] = true
	}

	pkgSlice, err := pkgList(excludedPackagesMap)
	if err != nil {
		return nil, err
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:              currentAPIKey,
		Logger:              log,
		PanicHandler:        func() {}, // we use our panic processor
		ProjectPackages:     pkgSlice,
		ReleaseStage:        cfg.ReleaseStage,
		NotifyReleaseStages: []string{production, staging},
		AppVersion:          version,
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
	return cli, nil
}

func emptyUUID() string {
	return new(uuid.UUID).String()
}

func metadata(addresses []string) bugsnag.MetaData {
	md := bugsnag.MetaData{accounts: {}}

	for k, v := range addresses {
		addr := v

		if !strings.HasPrefix(addr, "Ox") {
			addr = "0x" + addr
		}

		md[accounts][strconv.Itoa(k)] = addr
	}

	return md
}

func accEthAddresses(db *reform.DB) (addr []string) {
	accounts, err := db.SelectAllFrom(data.AccountTable, "")
	if err != nil || len(accounts) == 0 {
		return append(addr, defaultAccEth)
	}

	for k := range accounts {
		addr = append(addr, accounts[k].(*data.Account).EthAddr)
	}

	return addr
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
	addresses := accEthAddresses(c.db)

	var e error
	if sync {
		e = c.notifier.NotifySync(errors.New(err, skip),
			true, metadata(addresses))
	} else {
		e = c.notifier.Notify(errors.New(err, skip),
			metadata(addresses))
	}
	if e != nil {
		c.logger.Add("error", e).Debug("failed to send notify")
	}
}

// Enable returns true if client is enabled.
func (c *Client) Enable() bool {
	return c.enable
}

func (c *Client) allowed() bool {
	var setting data.Setting
	err := c.db.FindByPrimaryKeyTo(&setting, key)
	if err != nil {
		if err == sql.ErrNoRows {
			c.logger.Add("key", key).Warn("key is not exist" +
				" in Setting table")
		} else {
			c.logger.Add("key", key).Warn("failed to get key" +
				" from Setting table")
		}
		return false
	}
	val, err := strconv.ParseBool(setting.Value)
	if err != nil {
		c.logger.Add("key", key).Warn("key from Setting table" +
			" has an incorrect format")
		return false
	}
	return val
}

// PanicIgnore disables the PanicHunter.
func (c *Client) PanicIgnore() {
	panicIgnore = true
}

// PanicHunter catches panic, in case of an enabled reporter.
func PanicHunter() {
	if panicIgnore {
		return
	}

	if err := recover(); err != nil {
		if enable && notifier != nil {
			notifier.NotifySync(
				errors.New(err, 3), true,
				metadata([]string{defaultAccEth}))
		}
		panic(err)
	}
}
