package filter

import (
	"github.com/voronelf/logview/core"
	"strings"
)

type Contains struct {
	Field  string
	Substr string
}

var _ core.Filter = (*Contains)(nil)

func (f *Contains) Match(row core.Row) bool {
	rowValue, exists := row.Data[f.Field]
	if !exists {
		return false
	}
	rowValueString := toString(rowValue)
	for _, val := range strings.Split(f.Substr, "|") {
		if strings.Contains(rowValueString, val) {
			return true
		}
	}
	return false
}
