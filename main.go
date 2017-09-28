package main

import (
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"io"
	"os"
	"strings"
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

	if len(args) == 0 {
		args = append(args, "watch")
	} else if len(args) >= 1 {
		switch args[0] {
		case "-h", "--help", "--version":
			// do nothing
		case "help":
			args[0] = "--help"
		case "-v":
			args[0] = "--version"
		default:
			if strings.Contains(args[0], "-") {
				args = append([]string{"watch"}, args...)
			}
		}
	}
	c := cli.CLI{
		Name:       "logview",
		Version:    "0.9b",
		Args:       args,
		Commands:   commands,
		HelpWriter: helpWriter,
	}
	return c.Run()
}
