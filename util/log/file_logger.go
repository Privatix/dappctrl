package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// FileConfig is FileLogger configuration.
type FileConfig struct {
	*BaseConfig
	Prefix string
	UTC    bool
}

// NewFileConfig creates a new FileLogger configuration.
func NewFileConfig() *FileConfig {
	return &FileConfig{
		BaseConfig: NewBaseConfig(),
		Prefix:     "",
		UTC:        false,
	}
}

// LocationConfig is logs destination file configuration.
type LocationConfig struct {
	FilePrefix string
	Path       string
}

// NewLocationConfig creates a new FileLogger configuration.
func NewLocationConfig() *LocationConfig {
	return &LocationConfig{
		Path: "/var/log/",
	}
}

type fileLogger struct {
	*LoggerBase
	logger *log.Logger
}

// FileLoggerFile opens file.
func FileLoggerFile(conf *LocationConfig) (*os.File, error) {
	name := fmt.Sprintf(conf.Path+conf.FilePrefix+"-%s.log",
		time.Now().Format("2006-01-02"))
	return os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// NewFileLogger creates a new FileLogger.
func NewFileLogger(conf *FileConfig, out io.Writer) (Logger, error) {
	l := &fileLogger{}

	base, err := NewLoggerBase(conf.BaseConfig, l.log)
	if err != nil {
		return nil, err
	}

	flags := log.LstdFlags
	if conf.UTC {
		flags |= log.LUTC
	}

	l.logger = log.New(out, conf.Prefix, flags)
	l.LoggerBase = base

	return l, nil
}

// NewStderrLogger creates a new FileLogger for standard error stream.
func NewStderrLogger(conf *FileConfig) (Logger, error) {
	return NewFileLogger(conf, os.Stderr)
}

func (l *fileLogger) log(lvl Level, msg string,
	ctx map[string]interface{}, stack *string) error {
	var stack2 string
	if stack != nil {
		stack2 = "\n\n" + *stack + "\n"
	}

	l.logger.Printf("%-7s %s %v%s",
		strings.ToUpper(string(lvl)), msg, ctx, stack2)

	return nil
}
