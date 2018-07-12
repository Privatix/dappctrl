package log

import (
	"encoding/json"
	"time"

	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// Config is a database logger configuration.
type Config struct {
	*log.BaseConfig
}

// NewConfig creates a new database logger cofiguration.
func NewConfig() *Config {
	return &Config{
		BaseConfig: log.NewBaseConfig(),
	}
}

type dbLogger struct {
	*log.LoggerBase
	db *reform.DB
}

// NewLogger creates a new database logger.
func NewLogger(conf *Config, db *reform.DB) (log.Logger, error) {
	l := &dbLogger{db: db}

	base, err := log.NewLoggerBase(conf.BaseConfig, l.log)
	if err != nil {
		return nil, err
	}

	l.LoggerBase = base
	return l, nil
}

func (l *dbLogger) log(lvl log.Level, msg string,
	ctx map[string]interface{}, stack *string) error {
	ctxd, err := json.Marshal(ctx)
	if err != nil {
		return err
	}

	return l.db.Insert(&data.LogEvent{
		Time:    time.Now(),
		Level:   lvl,
		Message: msg,
		Context: ctxd,
		Stack:   stack,
	})
}
