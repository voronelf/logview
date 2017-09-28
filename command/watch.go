package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/voronelf/logview/core"
	"io"
	"strings"
)

type Watch struct {
	ShutdownCh    <-chan struct{}
	Stdin         io.Reader
	RowProvider   core.RowProvider   `inject:"RowProvider"`
	FilterFactory core.FilterFactory `inject:"FilterFactory"`
	Formatter     core.Formatter     `inject:"FormatterCliColor"`
	Ui            cli.Ui             `inject:"CliUi"`
	Settings      core.Settings      `inject:"Settings"`
}

var _ cli.Command = (*Watch)(nil)

func (c *Watch) Run(args []string) int {
	filePath, filterCondition, err := c.parseArgs(args)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	filter, err := c.FilterFactory.NewFilter(filterCondition)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	if filePath == "" {
		c.Ui.Output(messageWatchStdin(filterCondition))
		return c.watchStdin(filter)
	} else {
		c.Ui.Output(messageWatchFile(filePath, filterCondition))
		return c.watchFile(filePath, filter)
	}
}

func (c *Watch) parseArgs(args []string) (filePath, condition string, retErr error) {
	var tplName string
	cmdFlags := flag.NewFlagSet("watch", flag.ContinueOnError)
	cmdFlags.StringVar(&filePath, "f", "", "")
	cmdFlags.StringVar(&condition, "c", "", "")
	cmdFlags.StringVar(&tplName, "t", "", "")
	err := cmdFlags.Parse(args)
	if err != nil {
		return "", "", err
	}
	if tplName != "" {
		templates, err := c.Settings.GetTemplates()
		if err != nil {
			retErr = errors.New("template loading error: " + err.Error())
			return
		}
		tpl, ok := templates[tplName]
		if !ok {
			retErr = errors.New("template not found")
			return
		}
		if filePath == "" {
			tplFilePath, ok := tpl["f"]
			if ok {
				filePath = tplFilePath
			}
		}
		if condition == "" {
			tplCondition, ok := tpl["c"]
			if ok {
				condition = tplCondition
			}
		}
	}
	return
}

func (c *Watch) watchFile(filePath string, filter core.Filter) int {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	rowsChan, err := c.RowProvider.WatchFileChanges(ctx, filePath)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	for {
		select {
		case row, ok := <-rowsChan:
			if !ok {
				return 0
			}
			if row.Err != nil {
				c.Ui.Error(row.Err.Error())
				continue
			}
			if filter.Match(row) {
				c.Ui.Output(c.Formatter.Format(row))
			}
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (c *Watch) watchStdin(filter core.Filter) int {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	rowsChan, err := c.RowProvider.WatchOpenedStream(ctx, c.Stdin)
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	for {
		select {
		case row, ok := <-rowsChan:
			if !ok {
				return 0
			}
			if row.Err != nil {
				c.Ui.Error(row.Err.Error())
				continue
			}
			if filter.Match(row) {
				c.Ui.Output(c.Formatter.Format(row))
			}
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (*Watch) Synopsis() string {
	return "Default command. Subscribe on log file changes, analyze new rows and show rows matched by filter condition. Args: [-f filePath] [-c condition]"
}

func (*Watch) Help() string {
	text := `
Usage: logview watch [-f filePath] [-c condition]

    Subscribe on log file changes, analyze new rows and show rows matched by filter condition

Options:

    -f filePath    Log file path, if emtpy - used stdin
    -c condition   Filter condition. Contains one or more field checks.
                   Every field check is 'fieldName : fieldValue', where
                     fieldName  - name of field; can be wildcard with '*'
                     fieldValue - value of field, case insensitive;
                                  can be many values divided '|';
                                  every value can be wildcard with '*';
                                  every value can be negative, starts from '!'
                   Field checks are divided by logic operations: 'and', 'or'.
                   Also you can use brackets for prioritize operations.
`
	return strings.TrimSpace(text)
}

func messageWatchStdin(filterCondition string) string {
	return fmt.Sprintf("Watch with filter \"%s\"\n\n", filterCondition)
}

func messageWatchFile(filePath, filterCondition string) string {
	return fmt.Sprintf("Watch file \"%s\" with filter \"%s\"\n\n", filePath, filterCondition)
}
