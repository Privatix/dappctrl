package log

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/leekchan/timeutil"
)

// WriterConfig is an io.Writer based logger configuration.
type WriterConfig struct {
	*BaseConfig
	Prefix string
	UTC    bool
}

// NewWriterConfig creates a new io.Writer based logger configuration.
func NewWriterConfig() *WriterConfig {
	return &WriterConfig{
		BaseConfig: NewBaseConfig(),
		Prefix:     "",
		UTC:        false,
	}
}

type writerLogger struct {
	*LoggerBase
	logger *log.Logger
}

// NewWriterLogger creates a new io.Writer based logger.
func NewWriterLogger(conf *WriterConfig, out io.Writer) (Logger, error) {
	l := &writerLogger{}

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

// NewStderrLogger creates a new logger for standard error stream.
func NewStderrLogger(conf *WriterConfig) (Logger, error) {
	return NewWriterLogger(conf, os.Stderr)
}

// FileConfig is a file based logger configuration.
type FileConfig struct {
	*WriterConfig
	Filename string
	FileMode os.FileMode
}

// NewFileConfig creates a new file logger configuration.
func NewFileConfig() *FileConfig {
	return &FileConfig{
		WriterConfig: NewWriterConfig(),
		Filename:     "dappctrl-%Y-%m-%d.log",
		FileMode:     0644,
	}
}

// NewFileLogger creates a new file logger.
func NewFileLogger(conf *FileConfig) (Logger, io.Closer, error) {
	now := time.Now()
	if conf.UTC {
		now = now.UTC()
	}

	file, err := os.OpenFile(
		timeutil.Strftime(&now, conf.Filename),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, conf.FileMode)
	if err != nil {
		return nil, nil, err
	}

	logger, err := NewWriterLogger(conf.WriterConfig, file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return logger, file, nil
}

func (l *writerLogger) log(lvl Level, msg string,
	ctx map[string]interface{}, stack *string) error {
	var stack2 string
	if stack != nil {
		stack2 = "\n\n" + *stack + "\n"
	}

	l.logger.Printf("%-7s %s %v%s",
		strings.ToUpper(string(lvl)), msg, ctx, stack2)

	return nil
}
