package formatter

import (
	"github.com/fatih/color"
	wildcardPkg "github.com/ryanuber/go-glob"
	"github.com/voronelf/logview/core"
	"sort"
	"strings"
)

func NewCliColor() *cliColor {
	return &cliColor{}
}

type cliColor struct {
}

var _ core.Formatter = (*cliColor)(nil)

func (s *cliColor) Format(row core.Row, params core.FormatParams) string {
	clrAround := color.New(color.FgHiMagenta)
	clrAround.DisableColor()
	clrField := color.New(color.FgHiBlue)
	clrValue := color.New(color.FgGreen)
	clrAccentField := color.New(color.FgHiBlue)
	clrAccentValue := color.New(color.FgHiGreen)
	clrError := color.New(color.FgRed)

	divider := clrAround.Sprint("**********")
	text := divider + " " + s.formatHeader(row) + " " + divider + "\n"
	if row.Err == nil {
		fieldList := make([]string, 0, len(row.Data))
		if len(params.OutputFields) > 0 {
			for field := range row.Data {
				for _, wildcard := range params.OutputFields {
					if s.isMatchWildcard(wildcard, field) {
						fieldList = append(fieldList, field)
						break
					}
				}
			}
		} else {
			for field := range row.Data {
				fieldList = append(fieldList, field)
			}
		}
		sort.Strings(fieldList)
		for _, field := range fieldList {
			value, ok := row.Data[field]
			if !ok {
				continue
			}
			var accentFields []string
			if len(params.AccentFields) > 0 {
				accentFields = params.AccentFields
			} else {
				accentFields = []string{"message", "module", "request_system"}
			}
			accent := false
			for _, accentField := range accentFields {
				if field == accentField {
					accent = true
					break
				}
			}
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

func (*cliColor) isMatchWildcard(wildcard, value string) bool {
	if wildcard[0] == '!' {
		return !wildcardPkg.Glob(strings.ToLower(wildcard[1:]), strings.ToLower(value))
	} else {
		return wildcardPkg.Glob(strings.ToLower(wildcard), strings.ToLower(value))
	}
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
