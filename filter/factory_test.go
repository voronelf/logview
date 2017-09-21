package filter

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	lex "github.com/timtadh/lexmachine"
	"github.com/voronelf/logview/core"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func getRow() core.Row {
	row := core.Row{Data: map[string]interface{}{}}
	jsonData, err := ioutil.ReadFile("test/row.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonData, &row.Data)
	if err != nil {
		panic(err)
	}
	return row
}

// for debug
func test_showLexems(_ *testing.T) {
	condition := "intField = '123'"
	lexer, err := initLexer()
	if err != nil {
		fmt.Println(err)
		return
	}
	lower := strings.ToLower(condition)
	s, err := lexer.Scanner([]byte(lower))
	if err != nil {
		fmt.Println(err)
		return
	}
	for tok, err, eof := s.Next(); !eof; tok, err, eof = s.Next() {
		if err != nil {
			fmt.Println(err)
			return
		}
		token := tok.(*lex.Token)
		fmt.Printf("%-7v | %-10v\n", token.Type, string(token.Lexeme))
	}
	return
}

func TestFactory_NewFilter(t *testing.T) {
	row := getRow()
	cases := []struct {
		condition string
		expected  bool
	}{
		0:  {condition: "", expected: true},
		1:  {condition: " ", expected: true},
		2:  {condition: "*", expected: true},
		3:  {condition: " *", expected: true},
		4:  {condition: "* ", expected: true},
		5:  {condition: "intField: 123", expected: true},
		6:  {condition: "intField :123", expected: true},
		7:  {condition: "intField : 123", expected: true},
		8:  {condition: "intField: '123'", expected: true},
		9:  {condition: "intField: 12345", expected: false},
		10: {condition: "intField: 456|123|78", expected: true},
		11: {condition: "intField: !123", expected: false},
		12: {condition: "intField: !456", expected: true},
		13: {condition: "floatField: 56.78", expected: true},
		14: {condition: "strField: SomeString", expected: true},
		15: {condition: "StrField: someString", expected: true},
		16: {condition: "*STR*eld*: *some*", expected: true},
		17: {condition: "strField: wrongString", expected: false},
		18: {condition: "strLong: 'Many words in one string with spaces'", expected: true},
		19: {condition: "strLong: \"Many words in one string with spaces\"", expected: true},
		20: {condition: "strLong: *many*", expected: true},
		21: {condition: "strLong: *Words*", expected: true},
		22: {condition: "strLong: cucumber", expected: false},
		23: {condition: "strLong: !*Words*", expected: false},
		24: {condition: "strLong: !*cucumber*", expected: true},
		25: {condition: "intField: 123 and strField: someString", expected: true},
		26: {condition: "intField: 123 and strField: wrongString", expected: false},
		27: {condition: "intField: 123 or strField: someString", expected: true},
		28: {condition: "intField: 123 or strField: wrongString", expected: true},
		29: {condition: "intField:12345 and (strField:wrongString or floatField:56.78)", expected: false},
		30: {condition: "intField:'123' and (strField:wrongString or floatField:56.78)", expected: true},
		31: {condition: "(intField: '123' and strLong: '*Words*') or floatField: 56.78", expected: true},
		32: {condition: "intField: 78|654|123 and strLong: *Words*|*cucumber* and floatField: 56.78", expected: true},
	}
	for key, cs := range cases {
		t.Run(strconv.Itoa(key), func(t *testing.T) {
			factory := NewFactory()
			filter, err := factory.NewFilter(cs.condition)
			if assert.Nil(t, err, "Error not nil: %s, condition: '%s'", err, cs.condition) {
				assert.Equal(t, cs.expected, filter.Match(row))
			}
		})
	}
}

func TestFactory_NewFilter_SyntaxError(t *testing.T) {
	errorConditions := []string{
		"intField = 123 or floatField = 123 and strLong ~ Words",
		"intField or floatField = 123 and strLong ~ Words",
		"intField == 123",
		"intField floatField = 123",
		"intField ( floatField = 123 and strLong ~ Words)",
		"intField or ( floatField = 123 and strLong ~ Words",
	}
	for key, cond := range errorConditions {
		t.Run(strconv.Itoa(key), func(t *testing.T) {
			factory := NewFactory()
			_, err := factory.NewFilter(cond)
			assert.NotNil(t, err)
		})
	}
}
