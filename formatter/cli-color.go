package formatter

import (
	"github.com/fatih/color"
	"github.com/voronelf/logview/core"
	"sort"
)

func NewCliColor() *cliColor {
	return &cliColor{}
}

type cliColor struct {
}

var _ core.Formatter = (*cliColor)(nil)

func (s *cliColor) Format(row core.Row) string {
	clrAround := color.New(color.FgHiMagenta)
	clrAround.DisableColor()
	clrField := color.New(color.FgHiBlue)
	clrValue := color.New(color.FgGreen)
	clrAccentField := color.New(color.FgHiBlue)
	clrAccentValue := color.New(color.FgHiGreen)
	clrError := color.New(color.FgRed)

	accentFields := map[string]struct{}{
		"message":        {},
		"module":         {},
		"request_system": {},
	}

	divider := clrAround.Sprint("**********")
	text := divider + " " + s.formatHeader(row) + " " + divider + "\n"
	if row.Err == nil {
		fields := make([]string, 0, len(row.Data))
		for field := range row.Data {
			fields = append(fields, field)
		}
		sort.Strings(fields)
		for _, field := range fields {
			_, accent := accentFields[field]
			if accent {
				text += "   " + clrAccentField.Sprint(field) + clrAround.Sprint(": ") + clrAccentValue.Sprint(row.Data[field]) + "\n"
			} else {
				text += "   " + clrField.Sprint(field) + clrAround.Sprint(": ") + clrValue.Sprint(row.Data[field]) + "\n"
			}
		}
	} else {
		text += clrError.Sprintf("Logviewer row error: %v\n", row.Err)
	}
	text += divider
	return text
}

func (*cliColor) formatHeader(row core.Row) string {
	var c *color.Color
	level, ok := row.Data["level"].(string)
	if !ok {
		return "No level field"
	}
	switch level {
	case "debug":
		c = color.New(color.FgBlack, color.BgCyan)
	case "info":
		c = color.New(color.FgBlack, color.BgGreen)
	case "warn", "warning":
		c = color.New(color.FgBlack, color.BgHiYellow)
	case "error":
		c = color.New(color.FgBlack, color.BgRed)
	default:
		c = color.New(color.FgBlack, color.BgMagenta)
	}
	return c.Sprint("  Level: " + level + "  ")
}
