package util

import (
	"errors"
	gofmt "fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/privatix/dappctrl/report"
)

const logFilePerm = 0644

// Logger to log internal events.
type Logger struct {
	logger *log.Logger
	level  int
	rep    report.Reporter
	out    io.Writer
}

// LogConfig is a logger configuration.
type LogConfig struct {
	Level         string
	LogPath       string
	LogFilePrefix string
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

func createLogFile(prefix, path string) (*os.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, err
	}

	fileName := gofmt.Sprintf(prefix+"-%s.log",
		time.Now().Format("2006-01-02"))
	absolutePath := filepath.Join(path, fileName)

	return os.OpenFile(absolutePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		logFilePerm)
}

// NewLogger creates a new logger.
func NewLogger(conf *LogConfig) (*Logger, error) {
	lvl := parseLogLevel(conf.Level)
	if lvl < LogDebug || lvl > LogFatal {
		return nil, errors.New("bad log level")
	}

	logger := &Logger{level: lvl}

	if conf.LogPath != "" {
		file, err := createLogFile(conf.LogFilePrefix, conf.LogPath)
		if err != nil {
			logger.out = os.Stderr
		} else {
			logger.out = file
		}
	} else {
		logger.out = os.Stderr
	}

	logger.logger = log.New(logger.out, "", log.LstdFlags)

	return logger, nil
}

// GracefulStop closes the log file.
func (l *Logger) GracefulStop() {
	if file, ok := l.out.(*os.File); ok {
		if file != os.Stderr && file != os.Stdout &&
			file != os.Stdin {
			file.Close()
		}
	}
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
		// We cannot use os.Exit() not to ignore deferred calls.
		panic(gofmt.Sprintf(fmt, v...))
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

// Printf prints a log message.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Output(2, gofmt.Sprintf(format, v...))
}
