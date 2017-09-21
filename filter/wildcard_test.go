package filter

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestWildcard_Match(t *testing.T) {
	cases := []struct {
		field    string
		val      string
		expected bool
	}{
		{field: "strField", val: "someString", expected: true},
		{field: "strField", val: "other|someString|wrong", expected: true},
		{field: "strField", val: "otherString", expected: false},
		{field: "strField", val: "SomeStRinG", expected: false},
		{field: "strField", val: "some", expected: false},
		{field: "strField", val: "*", expected: true},
		{field: "strField", val: "some*", expected: true},
		{field: "strField", val: "*ing", expected: true},
		{field: "strField", val: "*some*", expected: true},
		{field: "strField", val: "*meStr*", expected: true},
		{field: "strField", val: "so*ing", expected: true},
		{field: "intField", val: "123", expected: true},
		{field: "intField", val: "12", expected: false},
		{field: "floatField", val: "56.78", expected: true},
		{field: "floatField", val: "56.780", expected: false},
		{field: "wrongField", val: "some", expected: false},
		{field: "wrongField", val: "*", expected: false},
		{field: "*", val: "123", expected: true},
		{field: "*", val: "111", expected: false},
		{field: "*", val: "words", expected: false},
		{field: "*Field", val: "56.78", expected: true},
		{field: "*", val: "*", expected: true},
		{field: "*wrong*", val: "*", expected: false},
		{field: "strField", val: "!someString", expected: false},
		{field: "strField", val: "!*meStr*", expected: false},
		{field: "strField", val: "!other", expected: true},
		{field: "strField", val: "wrong|!other", expected: true},
	}
	for k, cs := range cases {
		t.Run(strconv.Itoa(k), func(t *testing.T) {
			f := NewWildcard(cs.field, cs.val)
			assert.Equal(t, cs.expected, f.Match(getRow()))
		})
	}

}
