package main

import (
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"io"
	"os"
)

func main() {
	di := core.GetGlobalDIContainer()
	var ui cli.Ui = &cli.ConcurrentUi{Ui: &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr}}
	exitCode, err := doMain(di, ui, os.Stderr, os.Args[1:])
	if err != nil {
		ui.Error(err.Error())
	}
	os.Exit(exitCode)
}

func doMain(di *core.DIContainer, ui cli.Ui, helpWriter io.Writer, args []string) (int, error) {
	provideCommandsDependenciesInDI(di, ui)
	commands := getCommands(di)
	err := di.Populate()
	if err != nil {
		return 1, err
	}
	c := cli.CLI{
		Name:       "logview",
		Args:       args,
		Commands:   commands,
		HelpWriter: helpWriter,
	}
	return c.Run()
}
