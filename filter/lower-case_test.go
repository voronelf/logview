package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func TestLowerCase_Match(t *testing.T) {
	child := &core.MockFilter{}
	f := LowerCase{
		Child: child,
	}
	input := map[string]interface{}{
		"First":  "Value1",
		"seconD": float64(123.45),
	}
	expected := map[string]interface{}{
		"first":  "value1",
		"second": "123.45",
	}
	child.On("Match", core.Row{Data: expected}).Return(true)

	actual := f.Match(core.Row{Data: input})
	assert.True(t, actual)
	child.AssertExpectations(t)
}
