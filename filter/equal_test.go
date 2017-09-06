package filter

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestEqual_Match(t *testing.T) {
	cases := []struct {
		field    string
		val      string
		expected bool
	}{
		{field: "strField", val: "someString", expected: true},
		{field: "strField", val: "other|someString|wrong", expected: true},
		{field: "strField", val: "otherString", expected: false},
		{field: "strField", val: "some", expected: false},
		{field: "strField", val: "SomeStRinG", expected: false},
		{field: "intField", val: "123", expected: true},
		{field: "intField", val: "12", expected: false},
		{field: "floatField", val: "56.78", expected: true},
		{field: "floatField", val: "56.780", expected: false},
		{field: "wrongField", val: "some", expected: false},
	}
	for k, cs := range cases {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			f := Equal{
				Field: cs.field,
				Value: cs.val,
			}
			assert.Equal(t, cs.expected, f.Match(getRow()))
		})
	}
}
