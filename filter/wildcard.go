package filter

import (
	wildcardPkg "github.com/ryanuber/go-glob"
	"github.com/voronelf/logview/core"
	"strings"
)

func NewWildcard(field, value string) *wildcard {
	return &wildcard{
		field:           field,
		fieldIsWildcard: strings.Contains(field, "*"),
		values:          strings.Split(value, "|"),
	}
}

type wildcard struct {
	field           string
	fieldIsWildcard bool
	values          []string
}

var _ core.Filter = (*wildcard)(nil)

func (w *wildcard) Match(row core.Row) bool {
	if w.fieldIsWildcard {
		for rowField, rowValue := range row.Data {
			if !wildcardPkg.Glob(w.field, rowField) {
				continue
			}
			if w.matchRowValue(rowValue) {
				return true
			}
		}
	} else {
		rowValue, ok := row.Data[w.field]
		if ok && w.matchRowValue(rowValue) {
			return true
		}
	}
	return false
}

func (w *wildcard) matchRowValue(rowValue interface{}) bool {
	rowValueString := toString(rowValue)
	for _, val := range w.values {
		if val[0] != '!' {
			if wildcardPkg.Glob(val, rowValueString) {
				return true
			}
		} else {
			if !wildcardPkg.Glob(val[1:], rowValueString) {
				return true
			}
		}
	}
	return false
}
