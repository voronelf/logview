package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"strconv"
	"testing"
)

func TestAnd_Match(t *testing.T) {
	cases := []struct {
		left     bool
		right    bool
		expected bool
	}{
		{false, false, false},
		{false, true, false},
		{true, false, false},
		{true, true, true},
	}
	for key, cs := range cases {
		t.Run(strconv.Itoa(key), func(t *testing.T) {
			row := getRow()
			left := &core.MockFilter{}
			left.On("Match", row).Return(cs.left)
			right := &core.MockFilter{}
			right.On("Match", row).Return(cs.right)
			f := And{Left: left, Right: right}
			actual := f.Match(row)
			assert.Equal(t, cs.expected, actual)
		})
	}
}
