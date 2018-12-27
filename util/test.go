// +build !notest

package util

import (
	"encoding/json"
	"flag"
	"log"
	"testing"
)

// TestArgs is a test arguments.
type TestArgs struct {
	Conf    interface{}
	Verbose bool
}

// These are functions for shortening testing boilerplate.

// ReadTestArgs parses command line and reads arguments.
func ReadTestArgs(args *TestArgs) {
	fconfig := flag.String(
		"config", "dappctrl-test.config.json", "Configuration file")

	flag.BoolVar(&args.Verbose, "vv", false, "Verbose log output")

	flag.Parse()

	if err := ReadJSONFile(*fconfig, args.Conf); err != nil {
		log.Fatalf("failed to read configuration: %s\n", err)
	}
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

// TestUnmarshalJSON unmarshals a given JSON into a given object.
func TestUnmarshalJSON(t *testing.T, data []byte, v interface{}) {
	if data != nil {
		if err := json.Unmarshal(data, v); err != nil {
			t.Errorf("failed to unmarshal JSON: '%s' (%s)",
				err, Caller())
		}
	}
}
