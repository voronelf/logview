package filter

import (
	"github.com/voronelf/logview/core"
	"strings"
)

type Equal struct {
	Field string
	Value string
}

var _ core.Filter = (*Equal)(nil)

func (f *Equal) Match(row core.Row) bool {
	rowValue, exists := row.Data[f.Field]
	if !exists {
		return false
	}
	rowValueString := toString(rowValue)
	for _, val := range strings.Split(f.Value, "|") {
		if rowValueString == val {
			return true
		}
	}
	return false
}
