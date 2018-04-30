// +build !notest

package util

import (
	"flag"
	"log"
	"testing"
)

// These are functions for shortening testing boilerplate.

// ReadTestConfig parses command line and reads configuration.
func ReadTestConfig(conf interface{}) {
	fconfig := flag.String(
		"config", "dappctrl-test.config.json", "Configuration file")
	flag.Parse()

	if err := ReadJSONFile(*fconfig, conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}
}

// NewTestLogger creates a new logger.
func NewTestLogger(conf *LogConfig) *Logger {
	logger, err := NewLogger(conf)
	if err != nil {
		log.Fatalf("failed to create logger: %s\n", err)
	}
	return logger
}

// TestExpectResult compares two errors and fails a test if they don't match.
func TestExpectResult(t *testing.T, op string, expected, actual error) {
	sameContent := expected != nil && actual != nil &&
		expected.Error() == actual.Error()

	if expected != actual && !sameContent {
		t.Fatalf("unexpected '%s' result: expected '%v', returned "+
			"'%v' (%s)", op, expected, actual, Caller())
	}
}
