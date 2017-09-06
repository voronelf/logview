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
	var rowsCount int
	cmdFlags := flag.NewFlagSet("tail", flag.ContinueOnError)
	cmdFlags.StringVar(&filePath, "f", "", "")
	cmdFlags.StringVar(&filterCondition, "c", "", "")
	cmdFlags.IntVar(&rowsCount, "n", 0, "")
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

	filteredRowsCh, err := c.FileReader.ReadAll(filePath, filter)
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
	return "Analyze last n rows from log file and show rows matched by filter condition. Args: -f filePath [-c condition] [-n rowsCount]"
}

func (*Tail) Help() string {
	text := `
Usage: logview tail -f filePath [-c condition] [-n rowsCount]

    Analyze last n rows from log file and show rows matched by filter condition

Options:

    -f filePath    Log file path, required
    -c condition   Filter condition. Contains one or more field check.
                   Every field check is: 'fieldName operation fieldValue', where
                   operation is one of '=', '!=', '~', '!~'.
                   Field checks are divided by logic operations: 'and', 'or'.
                   Also you can use brackets.
    -n rowCount    Count last rows in file for analyzing
`
	return strings.TrimSpace(text)
}
