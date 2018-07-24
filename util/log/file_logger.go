package log

import (
	"io"
	"log"
	"os"
	"strings"
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

type fileLogger struct {
	*LoggerBase
	logger *log.Logger
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
