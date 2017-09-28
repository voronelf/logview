package command

import (
	"github.com/mitchellh/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/voronelf/logview/core"
	"strconv"
	"strings"
	"testing"
)

func newTplSaveForTest() *TplSave {
	return &TplSave{
		Settings: &core.MockSettings{},
		Ui:       &cli.MockUi{},
	}
}

func TestTplSave_Run(t *testing.T) {
	cases := []struct {
		args string
		name string
		tpl  core.Template
	}{
		{"-t tpl1 -f file", "tpl1", core.Template{"f": "file"}},
		{"-t tpl1 -a aaa -f file -c cond", "tpl1", core.Template{"a": "aaa", "f": "file", "c": "cond"}},
		{"-t tpl1", "tpl1", core.Template{}},
	}
	for i, cs := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cmd := newTplSaveForTest()
			cmd.Settings.(*core.MockSettings).On("SaveTemplate", cs.name, cs.tpl).Return(nil)
			cmd.Run(strings.Split(cs.args, " "))
			cmd.Settings.(*core.MockSettings).AssertExpectations(t)
		})
	}
}

func TestTplSave_Run_err(t *testing.T) {
	cmd := newTplSaveForTest()
	cmd.Run([]string{})
	cmd.Settings.(*core.MockSettings).AssertNotCalled(t, "SaveTemplate", mock.Anything, mock.Anything)
	assert.NotEmpty(t, cmd.Ui.(*cli.MockUi).ErrorWriter.String())
}
