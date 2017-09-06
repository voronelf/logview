package formatter

import (
	"github.com/fatih/color"
	"github.com/voronelf/logview/core"
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
		"message":        struct{}{},
		"module":         struct{}{},
		"request_system": struct{}{},
	}

	divider := clrAround.Sprint("**********")
	text := divider + " " + s.formatHeader(row) + " " + divider + "\n"
	if row.Err == nil {
		for field, value := range row.Data {
			_, accent := accentFields[field]
			if accent {
				text += "   " + clrAccentField.Sprint(field) + clrAround.Sprint(": ") + clrAccentValue.Sprint(value) + "\n"
			} else {
				text += "   " + clrField.Sprint(field) + clrAround.Sprint(": ") + clrValue.Sprint(value) + "\n"
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
