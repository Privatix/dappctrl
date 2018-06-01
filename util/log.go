package util

import (
	"errors"
	gofmt "fmt"
	"log"
	"os"
	"strings"

	"github.com/privatix/dappctrl/report"
)

// Logger to log internal events.
type Logger struct {
	logger *log.Logger
	level  int
	rep    report.Reporter
}

// LogConfig is a logger configuration.
type LogConfig struct {
	Level string
}

// Log levels.
const (
	LogDebug   = iota
	LogInfo    = iota
	LogWarning = iota
	LogError   = iota
	LogFatal   = iota
)

// NewLogConfig creates a default log configuration.
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level: logLevelStrs[LogInfo],
	}
}

var logLevelStrs = []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL"}

func parseLogLevel(lvl string) int {
	switch strings.ToUpper(lvl) {
	case logLevelStrs[LogDebug]:
		return LogDebug
	case logLevelStrs[LogInfo]:
		return LogInfo
	case logLevelStrs[LogWarning]:
		return LogWarning
	case logLevelStrs[LogError]:
		return LogError
	case logLevelStrs[LogFatal]:
		return LogFatal
	}

	return -1
}

// NewLogger creates a new logger.
func NewLogger(conf *LogConfig) (*Logger, error) {
	lvl := parseLogLevel(conf.Level)
	if lvl < LogDebug || lvl > LogFatal {
		return nil, errors.New("bad log level")
	}

	return &Logger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		level:  lvl,
	}, nil
}

// Reporter adds Reporter to the logger.
func (l *Logger) Reporter(reporter report.Reporter) {
	l.rep = reporter
}

// Log emits a log message.
func (l *Logger) Log(lvl int, fmt string, v ...interface{}) {
	if lvl < LogDebug || lvl > LogFatal || lvl < l.level {
		return
	}

	l.logger.Printf(logLevelStrs[lvl]+" "+fmt, v...)

	if l.rep != nil && l.rep.Enable() && lvl > LogWarning {
		e := gofmt.Errorf(logLevelStrs[lvl]+" "+fmt, v...)

		if lvl == LogError {
			l.rep.Notify(e, false, 4)
			return
		}
		l.rep.Notify(e, true, 4)
	}

	if lvl == LogFatal {
		os.Exit(1)
	}
}

// Debug emits a debugging message.
func (l *Logger) Debug(fmt string, v ...interface{}) {
	l.Log(LogDebug, fmt, v...)
}

// Info emits an information message.
func (l *Logger) Info(fmt string, v ...interface{}) {
	l.Log(LogInfo, fmt, v...)
}

// Warn emits an warning message.
func (l *Logger) Warn(fmt string, v ...interface{}) {
	l.Log(LogWarning, fmt, v...)
}

// Error emits an error message.
func (l *Logger) Error(fmt string, v ...interface{}) {
	l.Log(LogError, fmt, v...)
}

// Fatal emits a fatal message and exits with failure.
func (l *Logger) Fatal(fmt string, v ...interface{}) {
	l.Log(LogFatal, fmt, v...)
}

// Printf prints a log message
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Output(2, gofmt.Sprintf(format, v...))
}
