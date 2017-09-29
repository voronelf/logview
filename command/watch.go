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
	"time"
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
	filePath, filterCondition, formatParams, err := c.parseArgs(args)
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
		return c.watchStdin(filter, formatParams)
	} else {
		filePath = strings.Replace(filePath, "@today@", time.Now().UTC().Format("2006-01-02"), -1)
		c.Ui.Output(messageWatchFile(filePath, filterCondition))
		return c.watchFile(filePath, filter, formatParams)
	}
}

func (c *Watch) parseArgs(args []string) (filePath, condition string, formatParams core.FormatParams, err error) {
	formatParams = core.DefaultFormatParams()
	var tplName, showFields, accentFields string
	cmdFlags := flag.NewFlagSet("watch", flag.ContinueOnError)
	cmdFlags.StringVar(&filePath, "f", "", "")
	cmdFlags.StringVar(&condition, "c", "", "")
	cmdFlags.StringVar(&tplName, "t", "", "")
	cmdFlags.StringVar(&showFields, "o", "", "")
	cmdFlags.StringVar(&accentFields, "a", "", "")
	err = cmdFlags.Parse(args)
	if err != nil {
		return
	}
	if tplName != "" {
		templates, e := c.Settings.GetTemplates()
		if e != nil {
			err = errors.New("template loading error: " + e.Error())
			return
		}
		tpl, ok := templates[tplName]
		if !ok {
			err = errors.New("template not found")
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
		if showFields == "" {
			tplShowFields, ok := tpl["o"]
			if ok {
				showFields = tplShowFields
			}
		}
		if accentFields == "" {
			tplAccentFields, ok := tpl["a"]
			if ok {
				accentFields = tplAccentFields
			}
		}
	}
	if showFields != "" && showFields != "*" {
		fields := strings.Split(showFields, ",")
		for k, v := range fields {
			fields[k] = strings.TrimSpace(v)
		}
		formatParams.OutputFields = fields
	}
	if accentFields != "" && accentFields != "*" {
		fields := strings.Split(accentFields, ",")
		for k, v := range fields {
			fields[k] = strings.TrimSpace(v)
		}
		formatParams.AccentFields = fields
	}
	return
}

func (c *Watch) watchFile(filePath string, filter core.Filter, formatParams core.FormatParams) int {
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
				c.Ui.Output(c.Formatter.Format(row, formatParams))
			}
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (c *Watch) watchStdin(filter core.Filter, formatParams core.FormatParams) int {
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
				c.Ui.Output(c.Formatter.Format(row, formatParams))
			}
		case <-c.ShutdownCh:
			return 0
		}
	}
}

func (*Watch) Synopsis() string {
	return "Default command. Subscribe on log file changes, analyze new rows and show rows matched by filter condition. Args: [-f filePath] [-c condition]  [-o outputFields] [-a accentedFields]"
}

func (*Watch) Help() string {
	text := `
Usage: logview watch [-f filePath] [-c condition] [-o outputFields] [-a accentedFields]

    Subscribe on log file changes, analyze new rows and show rows matched by filter condition

Options:

    -f filePath    Log file path, if emtpy - used stdin. Substring '@today@' will be replace
                   to today date in format 2017-09-28.
    -c condition   Filter condition. Contains one or more field checks.
                   Every field check is 'fieldName : fieldValue', where
                     fieldName  - name of field; can be wildcard with '*'
                     fieldValue - value of field, case insensitive;
                                  can be many values divided '|';
                                  every value can be wildcard with '*';
                                  every value can be negative, starts from '!'
                   Field checks are divided by logic operations: 'and', 'or'.
                   Also you can use brackets for prioritize operations.
    -o fields      Comma-separated list of fields for output. Will show only this fields in that order.
                   Every field can be wildcard or negative wildcard (starts from !).
    -a fields      Comma-separated list of fields, which will show with high color.
`
	return strings.TrimSpace(text)
}

func messageWatchStdin(filterCondition string) string {
	return fmt.Sprintf("Watch with filter \"%s\"\n\n", filterCondition)
}

func messageWatchFile(filePath, filterCondition string) string {
	return fmt.Sprintf("Watch file \"%s\" with filter \"%s\"\n\n", filePath, filterCondition)
}
