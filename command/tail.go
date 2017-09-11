package command

import (
	"flag"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"strings"
)

type Tail struct {
	ShutdownCh    <-chan struct{}
	FileReader    core.FileReader    `inject:"FileReader"`
	FilterFactory core.FilterFactory `inject:"FilterFactory"`
	Formatter     core.Formatter     `inject:"FormatterCliColor"`
	Ui            cli.Ui             `inject:"CliUi"`
}

var _ cli.Command = (*Tail)(nil)

func (c *Tail) Run(args []string) int {
	var filePath, filterCondition string
	var bytesCount int64
	cmdFlags := flag.NewFlagSet("tail", flag.ContinueOnError)
	cmdFlags.StringVar(&filePath, "f", "", "")
	cmdFlags.StringVar(&filterCondition, "c", "", "")
	cmdFlags.Int64Var(&bytesCount, "b", 0, "")
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

	filteredRowsCh, err := c.FileReader.ReadTail(filePath, bytesCount, filter)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	for {
		select {
		case row, ok := <-filteredRowsCh:
			if !ok {
				return 0
			}
			if row.Err != nil {
				c.Ui.Error(err.Error())
				return 1
			}
			c.Ui.Output(c.Formatter.Format(row))
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (*Tail) Synopsis() string {
	return "Analyze last n rows from log file and show rows matched by filter condition. Args: -f filePath [-c condition] [-b bytes]"
}

func (*Tail) Help() string {
	text := `
Usage: logview tail -f filePath [-c condition] [-b bytes]

    Analyze last b bytes from log file and show rows matched by filter condition

Options:

    -f filePath    Log file path, required
    -c condition   Filter condition. Contains one or more field check.
                   Every field check is: 'fieldName operation fieldValue', where
                   operation is one of '=', '!=', '~', '!~'.
                   Field checks are divided by logic operations: 'and', 'or'.
                   Also you can use brackets.
    -b bytes       Count of bytes to last rows in file for analyzing
`
	return strings.TrimSpace(text)
}
