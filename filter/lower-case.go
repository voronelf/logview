package filter

import (
	"fmt"
	"github.com/voronelf/logview/core"
	"reflect"
	"strconv"
	"strings"
)

type LowerCase struct {
	Child core.Filter
}

var _ core.Filter = (*LowerCase)(nil)

func (f *LowerCase) Match(row core.Row) bool {
	r := row
	r.Data = make(map[string]interface{}, len(row.Data))
	for key, val := range row.Data {
		r.Data[strings.ToLower(key)] = strings.ToLower(toString(val))
	}
	return f.Child.Match(r)
}

func toString(val interface{}) string {
	switch reflect.TypeOf(val).Kind() {
	case reflect.String:
		return val.(string)
	case reflect.Float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	default:
		return fmt.Sprint(val)
	}
}
