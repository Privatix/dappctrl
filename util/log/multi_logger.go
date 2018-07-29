package log

type multiLogger []Logger

// NewMultiLogger creates a new multi-logger from given loggers.
func NewMultiLogger(loggers ...Logger) Logger {
	return multiLogger(loggers)
}

func (l multiLogger) Add(vars ...interface{}) Logger {
	res := make(multiLogger, len(l))
	for k, v := range l {
		res[k] = v.Add(vars...)
	}
	return res
}

func (l multiLogger) Log(lvl Level, msg string) {
	for _, v := range l {
		v.Log(lvl, msg)
	}
}

func (l multiLogger) Debug(msg string) { l.Log(Debug, msg) }
func (l multiLogger) Info(msg string)  { l.Log(Info, msg) }
func (l multiLogger) Warn(msg string)  { l.Log(Warning, msg) }
func (l multiLogger) Error(msg string) { l.Log(Error, msg) }
func (l multiLogger) Fatal(msg string) { l.Log(Fatal, msg) }
