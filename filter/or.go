package filter

import "github.com/voronelf/logview/core"

type Or struct {
	Left  core.Filter
	Right core.Filter
}

var _ core.Filter = (*Or)(nil)

func (f *Or) Match(row core.Row) bool {
	return f.Left.Match(row) || f.Right.Match(row)
}
