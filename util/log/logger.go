package log

import (
	"runtime/debug"
)

// Level is a log event severity.
type Level string

// Log severities.
const (
	Debug   Level = "debug"
	Info    Level = "info"
	Warning Level = "warning"
	Error   Level = "error"
	Fatal   Level = "fatal"
)

// Logger is an interface for loggers.
type Logger interface {
	Add(vars ...interface{}) Logger

	Log(lvl Level, msg string)

	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
}

// BaseConfig is a configuration for LoggerBase.
type BaseConfig struct {
	Level      Level
	StackLevel Level
}

// NewBaseConfig creates a new LoggerBase configuration.
func NewBaseConfig() *BaseConfig {
	return &BaseConfig{
		Level:      Info,
		StackLevel: Error,
	}
}

// LoggerFunc is a low-level logging function.
type LoggerFunc func(lvl Level, msg string,
	ctx map[string]interface{}, stack *string) error

// LoggerBase is a base logger implementation shared by many other loggers.
type LoggerBase struct {
	conf *BaseConfig
	log  LoggerFunc
	ctx  map[string]interface{}
}

var levelNums = map[Level]int{Debug: 0, Info: 1, Warning: 2, Error: 3, Fatal: 4}

// NewLoggerBase creates a new LoggerBase.
func NewLoggerBase(conf *BaseConfig, log LoggerFunc) (*LoggerBase, error) {
	if _, ok := levelNums[conf.Level]; !ok {
		return nil, ErrBadLevel
	}

	if _, ok := levelNums[conf.StackLevel]; !ok {
		return nil, ErrBadStackLevel
	}

	ctx := map[string]interface{}{}

	return &LoggerBase{conf: conf, log: log, ctx: ctx}, nil
}

// Add returns a new logger with added context variables. This function takes
// alternated key-value pairs (name1, val1, name1, val2, ...). The keys must be
// strings, but the values can be of any type.
func (l *LoggerBase) Add(vars ...interface{}) Logger {
	if len(vars)%2 == 1 {
		panic("bad number of arguments")
	}

	ctx := map[string]interface{}{}

	for k, v := range l.ctx {
		ctx[k] = v
	}

	num := len(vars) / 2
	for i := 0; i < num; i++ {
		name, ok := vars[i*2].(string)
		if !ok {
			panic("non-string variable name")
		}
		ctx[name] = vars[i*2+1]
	}

	return &LoggerBase{conf: l.conf, log: l.log, ctx: ctx}
}

// Log adds a new log message with a given severity level.
func (l *LoggerBase) Log(lvl Level, msg string) {
	lvln, ok := levelNums[lvl]
	if !ok {
		panic("bad log level")
	}

	if lvln < levelNums[l.conf.Level] {
		return
	}

	var stack *string
	if lvln >= levelNums[l.conf.StackLevel] {
		tmp := string(debug.Stack())
		stack = &tmp
	}

	if err := l.log(lvl, msg, l.ctx, stack); err != nil {
		panic("failed to log: " + err.Error())
	}

	if lvl == Fatal {
		panic("fatal log event")
	}
}

// Debug adds a new debug message.
func (l *LoggerBase) Debug(msg string) { l.Log(Debug, msg) }

// Info adds a new info message.
func (l *LoggerBase) Info(msg string) { l.Log(Info, msg) }

// Warn adds a new warning message.
func (l *LoggerBase) Warn(msg string) { l.Log(Warning, msg) }

// Error adds a new error message.
func (l *LoggerBase) Error(msg string) { l.Log(Error, msg) }

// Fatal adds a new fatal message and then panics.
func (l *LoggerBase) Fatal(msg string) { l.Log(Fatal, msg) }
