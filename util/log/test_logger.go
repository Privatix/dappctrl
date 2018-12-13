package log

// NewTestLogger creates a new test logger.
func NewTestLogger(conf *WriterConfig, verbose bool) (Logger, error) {
	if verbose {
		return NewStderrLogger(conf)
	}
	return NewMultiLogger(), nil
}
