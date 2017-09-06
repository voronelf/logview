package filter

import "github.com/voronelf/logview/core"

type Not struct {
	Child core.Filter
}

var _ core.Filter = (*Not)(nil)

func (f *Not) Match(row core.Row) bool {
	return !f.Child.Match(row)
}
