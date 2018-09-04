package version

import (
	"fmt"
	"os"
)

const (
	undefined = "undefined"
)

// Print prints version and completes the program.
func Print(run bool, commit, version string) {
	if run {
		fmt.Println(Message(commit, version))
		os.Exit(0)
	}
}

// Message returns version with format
// `version (first seven characters of commit hash)`.
func Message(commit, version string) string {
	var c string
	var v string

	if commit == "" {
		c = undefined
	} else if len(commit) > 7 {
		c = commit[:7]
	} else {
		c = commit
	}

	if version == "" {
		v = undefined
	} else {
		v = version
	}

	return fmt.Sprintf("%s (%s)", v, c)
}
