package filter

import "github.com/voronelf/logview/core"

type And struct {
	Left  core.Filter
	Right core.Filter
}

var _ core.Filter = (*And)(nil)

func (f *And) Match(row core.Row) bool {
	return f.Left.Match(row) && f.Right.Match(row)
}
