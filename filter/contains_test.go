package filter

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestContains_Match(t *testing.T) {
	cases := []struct {
		field    string
		substr   string
		expected bool
	}{
		{field: "strLong", substr: "Many words", expected: true},
		{field: "strLong", substr: "many words", expected: false},
		{field: "strlong", substr: "Many words", expected: false},
		{field: "strLong", substr: "Many, words", expected: false},
		{field: "strLong", substr: "wrong|words|other", expected: true},
		{field: "strField", substr: "someString", expected: true},
		{field: "strField", substr: "some", expected: true},
		{field: "strField", substr: "String", expected: true},
		{field: "strField", substr: "other", expected: false},
		{field: "intField", substr: "12", expected: true},
		{field: "floatField", substr: "6.7", expected: true},
		{field: "floatField", substr: "56.780", expected: false},
		{field: "wrongField", substr: "some", expected: false},
	}
	for k, cs := range cases {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			f := Contains{
				Field:  cs.field,
				Substr: cs.substr,
			}
			assert.Equal(t, cs.expected, f.Match(getRow()))
		})
	}
}
