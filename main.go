package main

import (
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/command"
	"github.com/voronelf/logview/core"
	"github.com/voronelf/logview/filter"
	"github.com/voronelf/logview/formatter"
	"github.com/voronelf/logview/file"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	di := core.GetDIContainer()

	var ui cli.Ui = &cli.ConcurrentUi{
		Ui: &cli.BasicUi{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
	di.Provide("CliUi", ui)

	var obs core.Observer = file.NewObserver()
	di.Provide("Observer", obs)

	var ff core.FilterFactory = filter.NewFactory()
	di.Provide("FilterFactory", ff)

	var f core.Formatter = formatter.NewCliColor()
	di.Provide("FormatterCliColor", f)
}

func main() {
	di := core.GetDIContainer()
	commands := getCommands(di)
	err := di.Populate()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	c := cli.CLI{
		Name:     "logview",
		Args:     os.Args[1:],
		Commands: commands,
	}

	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func getCommands(di *core.DIContainer) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"watch": newCmdFactory(di, &command.Watch{ShutdownCh: getShutdownCh()}),
		"tail":  newCmdFactory(di, &command.Tail{ShutdownCh: getShutdownCh()}),
	}
}

func newCmdFactory(di *core.DIContainer, cmd cli.Command) cli.CommandFactory {
	di.Provide("", cmd)
	return func() (cli.Command, error) { return cmd, nil }
}

var shutdownCh chan struct{}

// getShutdownCh returns a channel that will be closed for every interrupt or SIGTERM received.
func getShutdownCh() <-chan struct{} {
	if shutdownCh == nil {
		shutdownCh = make(chan struct{})
		go func() {
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
			for {
				<-signalCh
				close(shutdownCh)
			}
		}()
	}
	return shutdownCh
}
