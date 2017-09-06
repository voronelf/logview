package filter

import (
	"github.com/stretchr/testify/assert"
	"github.com/voronelf/logview/core"
	"testing"
)

func TestAll_Match(t *testing.T) {
	f := All{}
	row := core.Row{Data: map[string]interface{}{}}
	assert.True(t, f.Match(row))
}
