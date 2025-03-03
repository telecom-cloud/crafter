package main

import (
	"os"

	"github.com/telecom-cloud/crafter/cmd/cft/app"
	"github.com/telecom-cloud/crafter/pkg/util/logs"
)

func main() {
	// run in plugin mode
	app.PluginMode()

	// run in normal mode
	NormalMode()
}

func NormalMode() {
	defer func() {
		logs.Flush()
	}()

	cmd := app.NewCommand()
	err := cmd.Run(os.Args)
	if err != nil {
		logs.Errorf("%v\n", err)
	}
}
