package main

import (
	"fmt"
	"os"

	"github.com/privatix/dappctrl/assemble"
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/report/bugsnag"
)

func main() {
	if err := data.ExecuteCommand(os.Args[1:]); err != nil {
		panic(fmt.Sprintf("failed to execute command: %s", err))
	}

	fatal := make(chan error)
	defer bugsnag.PanicHunter()

	assemble.RunApp(fatal)
}
