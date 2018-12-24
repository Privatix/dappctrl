package assemble

import (
	abill "github.com/privatix/dappctrl/agent/bill"
	cbill "github.com/privatix/dappctrl/client/bill"
	"github.com/privatix/dappctrl/country"
	"github.com/privatix/dappctrl/data"
	dblog "github.com/privatix/dappctrl/data/log"
	"github.com/privatix/dappctrl/eth"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages/ept"
	"github.com/privatix/dappctrl/monitor"
	"github.com/privatix/dappctrl/pay"
	"github.com/privatix/dappctrl/proc"
	"github.com/privatix/dappctrl/proc/looper"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/report/bugsnag"
	rlog "github.com/privatix/dappctrl/report/log"
	"github.com/privatix/dappctrl/somc"
	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/rpcsrv"
)

// Config is application config.
type Config struct {
	AgentDirectSOMC *somc.DirectAgentConfig
	AgentMonitor    *abill.Config
	AgentTorSOMC    *somc.TorAgentConfig
	BlockMonitor    *monitor.Config
	ClientMonitor   *cbill.Config
	ClientTorSOMC   *somc.TorClientConfig
	Country         *country.Config
	DB              *data.DBConfig
	DBLog           *dblog.Config
	ReportLog       *rlog.Config
	EptMsg          *ept.Config
	Eth             *eth.Config
	FileLog         *log.FileConfig
	Gas             *worker.GasConf
	Job             *job.Config
	Looper          *looper.Config
	PayServer       *pay.Config
	PayAddress      string
	Proc            *proc.Config
	Report          *bugsnag.Config
	Role            string
	Sess            *rpcsrv.Config
	SOMCServer      *rpcsrv.Config
	StaticPassword  string
	UI              *rpcsrv.Config
}

// NewConfig returns application config with default values.
func NewConfig() *Config {
	return &Config{
		AgentDirectSOMC: somc.NewDirectAgentConfig(),
		AgentMonitor:    abill.NewConfig(),
		AgentTorSOMC:    somc.NewTorAgentConfig(),
		BlockMonitor:    monitor.NewConfig(),
		ClientMonitor:   cbill.NewConfig(),
		ClientTorSOMC:   somc.NewTorClientConfig(),
		Country:         country.NewConfig(),
		DB:              data.NewDBConfig(),
		DBLog:           dblog.NewConfig(),
		Looper:          looper.NewConfig(),
		ReportLog:       rlog.NewConfig(),
		EptMsg:          ept.NewConfig(),
		Eth:             eth.NewConfig(),
		FileLog:         log.NewFileConfig(),
		Job:             job.NewConfig(),
		PayServer:       pay.NewConfig(),
		Proc:            proc.NewConfig(),
		Report:          bugsnag.NewConfig(),
		Sess:            rpcsrv.NewConfig(),
		SOMCServer:      rpcsrv.NewConfig(),
		UI:              rpcsrv.NewConfig(),
	}
}
