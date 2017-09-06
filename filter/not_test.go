package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"strconv"
	"testing"
)

func TestNot_Match(t *testing.T) {
	cases := []struct {
		in       bool
		expected bool
	}{
		{true, false},
		{false, true},
	}
	for key, cs := range cases {
		t.Run(strconv.Itoa(key), func(t *testing.T) {
			row := getRow()
			child := &core.MockFilter{}
			child.On("Match", row).Return(cs.in).Once()
			f := Not{Child: child}
			match := f.Match(row)
			assert.Equal(t, cs.expected, match)
			child.AssertExpectations(t)
		})
	}
}
