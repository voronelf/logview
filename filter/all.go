package filter

import (
	"github.com/voronelf/logview/core"
)

type All struct {
}

var _ core.Filter = (*All)(nil)

func (*All) Match(row core.Row) bool {
	return true
}
