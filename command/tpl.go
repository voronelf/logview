package command

import "github.com/mitchellh/cli"

type Tpl struct {
}

var _ cli.Command = (*Tpl)(nil)

func (*Tpl) Run(args []string) int {
	return cli.RunResultHelp
}

func (*Tpl) Help() string {
	return "Working with templates of parameters. For more info see subcommands"
}

func (*Tpl) Synopsis() string {
	return "Working with templates of parameters. For more info see subcommands"
}
