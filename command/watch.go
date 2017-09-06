package command

import (
	"context"
	"flag"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"strings"
)

type Watch struct {
	ShutdownCh    <-chan struct{}
	Observer      core.Observer      `inject:"Observer"`
	FilterFactory core.FilterFactory `inject:"FilterFactory"`
	Formatter     core.Formatter     `inject:"FormatterCliColor"`
	Ui            cli.Ui             `inject:"CliUi"`
}

var _ cli.Command = (*Watch)(nil)

func (c *Watch) Run(args []string) int {
	var filePath, filterCondition string
	cmdFlags := flag.NewFlagSet("watch", flag.ContinueOnError)
	cmdFlags.StringVar(&filePath, "f", "", "")
	cmdFlags.StringVar(&filterCondition, "c", "", "")
	err := cmdFlags.Parse(args)
	if err != nil {
		return cli.RunResultHelp
	}
	if filePath == "" {
		return cli.RunResultHelp
	}

	filter, err := c.FilterFactory.NewFilter(filterCondition)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	s, err := c.Observer.Subscribe(ctx, filePath, filter)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	c.Ui.Output("Watch file " + filePath + " with filter \"" + filterCondition + "\"")
	for {
		select {
		case row := <-s.Channel:
			c.Ui.Output(c.Formatter.Format(row))
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (*Watch) Synopsis() string {
	return "Subscribe on log file changes, analize new rows and show rows matched by filter condition. Args: -f filePath [-c condition]"
}

func (*Watch) Help() string {
	text := `
Usage: logview watch -f filePath [-c condition]

    Subscribe on log file changes, analize new rows and show rows matched by filter condition

Options:

    -f filePath    Log file path, required
    -c condition   Filter condition. Contains one or more field check.
                   Every field check is: 'fieldName operation fieldValue', where
                   operation is one of '=', '!=', '~', '!~'.
                   Field checks are divided by logic operations: 'and', 'or'.
                   Also you can use brackets.
`
	return strings.TrimSpace(text)
}
