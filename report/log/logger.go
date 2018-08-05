package log

import (
	"fmt"

	"github.com/privatix/dappctrl/report"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a reporter logger configuration.
type Config struct {
	*log.BaseConfig
}

// NewConfig creates a new reporter logger configuration.
func NewConfig() *Config {
	return &Config{
		BaseConfig: log.NewBaseConfig(),
	}
}

type reportLogger struct {
	*log.LoggerBase
	rep    report.Reporter
	logger log.Logger // logger for internal messages
}

// NewLogger creates a new reporter logger.
func NewLogger(conf *Config) (log.Logger, error) {
	l := &reportLogger{}

	base, err := log.NewLoggerBase(conf.BaseConfig, l.log)
	if err != nil {
		return nil, err
	}

	l.LoggerBase = base
	return l, nil
}

// Reporter adds Reporter to the logger.
func (l *reportLogger) Reporter(reporter report.Reporter) {
	l.rep = reporter
}

// Reporter adds a global logger to the logger.
func (l *reportLogger) Logger(logger log.Logger) {
	l.logger = logger
}

func (l *reportLogger) log(lvl log.Level, msg string,
	ctx map[string]interface{}, stack *string) error {
	if l.rep != nil && l.rep.Enable() &&
		(lvl == log.Error || lvl == log.Fatal) {
		e := fmt.Errorf(string(lvl) + " " + msg)

		if lvl == log.Error {
			l.rep.Notify(e, false, 2)
			return nil
		}
		l.rep.Notify(e, true, 2)
		l.rep.PanicIgnore()
	}
	return nil
}

// Printf logs internal messages from a reporter.
func (l *reportLogger) Printf(format string, v ...interface{}) {
	if l.logger != nil {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}
