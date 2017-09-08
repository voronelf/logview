package main

import (
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func TestProvidingCommandsDependencies(t *testing.T) {
	di := core.NewDIContainer()
	ui := &cli.MockUi{}
	provideCommandsDependenciesInDI(di, ui)
	getCommands(di)
	err := di.Populate()
	assert.Nil(t, err)
}
