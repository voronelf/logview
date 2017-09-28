package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"strings"
)

type TplList struct {
	Ui       cli.Ui        `inject:"CliUi"`
	Settings core.Settings `inject:"Settings"`
}

var _ cli.Command = (*TplList)(nil)

func (t *TplList) Run(args []string) int {
	templates, err := t.Settings.GetTemplates()
	if err != nil {
		t.Ui.Error("Error: " + err.Error())
		return 1
	}
	for name, tpl := range templates {
		desc := ""
		for flag, value := range tpl {
			desc += "-" + flag + " \"" + value + "\""
		}
		t.Ui.Output(fmt.Sprintf("\t%s :\t%s\n", name, desc))
	}
	return 0
}

func (*TplList) Synopsis() string {
	return "Show list of saved templates"
}

func (*TplList) Help() string {
	text := `
Usage: logview tpl list

    Show list of saved templates
`
	return strings.TrimSpace(text)
}
