package command

import (
	"errors"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"strings"
)

type TplSave struct {
	Settings core.Settings `inject:"Settings"`
	Ui       cli.Ui        `inject:"CliUi"`
}

var _ cli.Command = (*TplSave)(nil)

func (t *TplSave) Run(args []string) int {
	tplName, tpl, err := t.parseArgs(args)
	if err != nil {
		t.Ui.Error("Parse args error: " + err.Error())
		return 1
	}
	if tplName == "" {
		t.Ui.Error("Must be -t parameter")
		return 1
	}
	err = t.Settings.SaveTemplate(tplName, tpl)
	if err != nil {
		t.Ui.Error("Save error: " + err.Error())
		return 1
	}
	return 0
}

func (*TplSave) parseArgs(args []string) (tplName string, tpl core.Template, err error) {
	tpl = make(core.Template, 2)
	l := len(args)
	for i := 0; i < l; i++ {
		s := args[i]
		if len(s) == 0 || s[0] != '-' || len(s) == 1 {
			err = errors.New("wrong flag: " + s)
			return
		}
		numMinuses := 1
		if s[1] == '-' {
			numMinuses++
			if len(s) == 2 {
				err = errors.New("wrong flag: " + s)
				return
			}
		}
		name := s[numMinuses:]
		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			err = errors.New("bad flag syntax: " + s)
			return
		}
		// it's a flag. does it have an argument?
		valueFound := false
		value := ""
		for i := 1; i < len(name); i++ { // equals cannot be first
			if name[i] == '=' {
				value = name[i+1:]
				valueFound = true
				name = name[0:i]
				break
			}
		}
		if !valueFound && i+1 < l {
			i = i + 1
			value = args[i]
		}
		if name == "t" {
			tplName = value
		} else {
			tpl[name] = value
		}
	}
	return
}

func (*TplSave) Synopsis() string {
	return "Save all flags as template with name in -t parameter. Params: -t name [any other flags to save in template]"
}

func (*TplSave) Help() string {
	text := `
Usage: logview tpl save -t name ...params to save...

    Save all flags as template with name in -t parameter

Options:
    -t name  Name for saved template
    ...      Any other params will be saved as template content. If content is empty - template will be deleted.
`
	return strings.TrimSpace(text)
}
