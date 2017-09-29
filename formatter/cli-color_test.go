package formatter

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestCliColor_isMatchWildcard(t *testing.T) {
	cases := []struct {
		wld string
		val string
		exp bool
	}{
		{"abcdef", "abcdef", true},
		{"AbCdeF", "aBcdef", true},
		{"a*f", "abcdef", true},
		{"*c*e*", "abcdef", true},
		{"*c*e*", "ABCDEF", true},
		{"*ccc*", "abcdef", false},
		{"!*ccc*", "abcdef", true},
		{"!a*f", "abcdef", false},
		{"!abcdef", "abcdef", false},
		{"*", "abcdef", true},
		{"!*", "abcdef", false},
	}
	for i, cs := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := NewCliColor().isMatchWildcard(cs.wld, cs.val)
			assert.Equal(t, cs.exp, actual)
		})
	}

}
