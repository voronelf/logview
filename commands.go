package main

import (
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/command"
	"github.com/voronelf/logview/core"
	"github.com/voronelf/logview/filter"
	"github.com/voronelf/logview/formatter"
	"github.com/voronelf/logview/provider"
	"os"
	"os/signal"
	"syscall"
)

func provideCommandsDependenciesInDI(di *core.DIContainer, ui cli.Ui) {
	di.Provide("CliUi", ui)

	var rowProvider core.RowProvider = provider.NewRowProvider()
	di.Provide("RowProvider", rowProvider)

	var factory core.FilterFactory = filter.NewFactory()
	di.Provide("FilterFactory", factory)

	var f core.Formatter = formatter.NewCliColor()
	di.Provide("FormatterCliColor", f)
}

func getCommands(di *core.DIContainer) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"watch": newCmdFactory(di, &command.Watch{ShutdownCh: getShutdownCh(), Stdin: os.Stdin}),
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
