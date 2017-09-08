package main

import (
	"bytes"
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func TestDoMain_help(t *testing.T) {
	di := core.NewDIContainer()
	ui := &cli.MockUi{}
	helpWriter := &bytes.Buffer{}

	exitCode, err := doMain(di, ui, helpWriter, []string{"--help"})
	assert.Nil(t, err)
	assert.Equal(t, 1, exitCode)
	assert.Empty(t, ui.OutputWriter)
	assert.NotEmpty(t, helpWriter)
}
