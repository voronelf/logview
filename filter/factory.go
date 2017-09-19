package filter

import (
	"fmt"
	"github.com/pkg/errors"
	lex "github.com/timtadh/lexmachine"
	"github.com/timtadh/lexmachine/machines"
	"github.com/voronelf/logview/core"
	"strings"
)

func NewFactory() *factory {
	return &factory{}
}

type factory struct {
}

func (*factory) NewFilter(condition string) (core.Filter, error) {
	cleanedCondition := strings.ToLower(strings.TrimSpace(condition))
	if cleanedCondition == "" || cleanedCondition == "*" {
		return &All{}, nil
	}
	lexer, err := initLexer()
	if err != nil {
		return nil, err
	}
	scanner, err := lexer.Scanner([]byte(cleanedCondition))
	if err != nil {
		return nil, err
	}
	node, err := createFiltersTree(scanner, false)
	if err != nil {
		return nil, err
	}
	return &LowerCase{Child: node}, nil
}

const (
	opEqual       = "="
	opNotEqual    = "!="
	opContains    = "~"
	opNotContains = "!~"
	opAnd         = "and"
	opOr          = "or"
)

const (
	typeString int = iota
	typeFieldOperation
	typeCondOperation
	typeOpenBracket
	typeCloseBracket
)

func initLexer() (*lex.Lexer, error) {
	lexer := lex.NewLexer()
	lexer.Add([]byte("([a-z]|[0-9]|_|\\-|\\.|\\|)+"), analyzeString)
	lexer.Add([]byte("\\'"), giveStringBetweenQuotes('\''))
	lexer.Add([]byte("\\\""), giveStringBetweenQuotes('"'))
	lexer.Add([]byte("\\=|\\!\\=|\\~|\\!\\~"), token(typeFieldOperation))
	lexer.Add([]byte("\\("), token(typeOpenBracket))
	lexer.Add([]byte("\\)"), token(typeCloseBracket))
	lexer.Add([]byte("( |\t|\n|\r)+"), skip)

	err := lexer.Compile()
	if err != nil {
		return nil, err
	}
	return lexer, nil
}

func token(tokenType int) lex.Action {
	return func(s *lex.Scanner, m *machines.Match) (interface{}, error) {
		return s.Token(tokenType, string(m.Bytes), m), nil
	}
}

func analyzeString(s *lex.Scanner, m *machines.Match) (interface{}, error) {
	tokenType := typeString
	strMatch := string(m.Bytes)
	if strMatch == opAnd || strMatch == opOr {
		tokenType = typeCondOperation
	}
	return s.Token(tokenType, strMatch, m), nil
}

func skip(_ *lex.Scanner, _ *machines.Match) (interface{}, error) {
	return nil, nil
}

func giveStringBetweenQuotes(quoteSymbol byte) lex.Action {
	return func(scan *lex.Scanner, match *machines.Match) (interface{}, error) {
		str := make([]byte, 0, 10)
		match.EndLine = match.StartLine
		match.EndColumn = match.StartColumn
		for tc := scan.TC; tc < len(scan.Text); tc++ {
			match.EndColumn += 1
			if scan.Text[tc] == '\n' {
				match.EndLine += 1
			}
			if scan.Text[tc] == quoteSymbol {
				match.TC = scan.TC
				scan.TC = tc + 1
				match.Bytes = str
				return token(typeString)(scan, match)
			}
			str = append(str, scan.Text[tc])
		}
		return nil,
			fmt.Errorf("unclosed string value, started at %d, (%d, %d)",
				match.TC, match.StartLine, match.StartColumn)
	}
}

func createFiltersTree(s *lex.Scanner, expectCloseBracket bool) (core.Filter, error) {
	var root core.Filter
	var logicOperationToken *lex.Token
	for i := 0; i < 100; i++ {
		firstToken, err := requiredTokenFromScanner(s)
		if err != nil {
			return nil, err
		}
		var node core.Filter
		switch firstToken.Type {
		case typeString:
			node, err = createFieldOperation(s, firstToken)
			if err != nil {
				return nil, err
			}
		case typeOpenBracket:
			node, err = createFiltersTree(s, true)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("Unexpected token type for '%s'. Expected field name or open bracket", string(firstToken.Lexeme))
		}
		if logicOperationToken != nil {
			root, err = createLogicOperation(logicOperationToken, root, node)
			if err != nil {
				return nil, err
			}
		} else {
			root = node
		}

		token, eof, err := tokenFromScanner(s)
		if err != nil {
			return nil, err
		}
		if eof {
			if expectCloseBracket {
				return nil, errors.Errorf("Not found close bracket")
			} else {
				return root, nil
			}
		}
		switch token.Type {
		case typeCloseBracket:
			if expectCloseBracket {
				return root, nil
			}
			return nil, errors.Errorf("Found close bracket without open bracket")
		case typeCondOperation:
			if logicOperationToken != nil && string(token.Lexeme) != string(logicOperationToken.Lexeme) {
				return nil, errors.Errorf("Different logic operations in one condition. Priorities is not defined. Let's use brackets")
			}
			logicOperationToken = token
			continue
		default:
			return nil, errors.Errorf("Unexpected token '%s'. Expected logic operation or close bracket")
		}
	}
	return nil, errors.New("Many logical operations in condition")
}

func requiredTokenFromScanner(s *lex.Scanner) (*lex.Token, error) {
	tok, err, eof := s.Next()
	if err != nil {
		return nil, err
	}
	if eof {
		return nil, errors.New("Unexpected eof")
	}
	return tok.(*lex.Token), nil
}

func tokenFromScanner(s *lex.Scanner) (*lex.Token, bool, error) {
	tok, err, eof := s.Next()
	if err != nil || eof {
		return nil, eof, err
	}
	return tok.(*lex.Token), eof, nil
}

func createFieldOperation(s *lex.Scanner, fieldToken *lex.Token) (core.Filter, error) {
	operationToken, err := requiredTokenFromScanner(s)
	if err != nil {
		return nil, err
	}
	operation := string(operationToken.Lexeme)
	if operationToken.Type != typeFieldOperation {
		return nil, errors.Errorf("Unexpected operation type '%s'", operation)
	}
	fieldValueToken, err := requiredTokenFromScanner(s)
	if err != nil {
		return nil, err
	}
	if fieldValueToken.Type != typeString {
		return nil, errors.Errorf("Unexpected token type for field value: '%s'", string(fieldValueToken.Lexeme))
	}
	switch operation {
	case opEqual:
		return &Equal{
			Field: string(fieldToken.Lexeme),
			Value: string(fieldValueToken.Lexeme),
		}, nil
	case opNotEqual:
		return &Not{
			Child: &Equal{
				Field: string(fieldToken.Lexeme),
				Value: string(fieldValueToken.Lexeme),
			},
		}, nil
	case opContains:
		return &Contains{
			Field:  string(fieldToken.Lexeme),
			Substr: string(fieldValueToken.Lexeme),
		}, nil
	case opNotContains:
		return &Not{
			Child: &Contains{
				Field:  string(fieldToken.Lexeme),
				Substr: string(fieldValueToken.Lexeme),
			},
		}, nil
	default:
		return nil, errors.Errorf("Unknown operation '%s'", operation)
	}
}

func createLogicOperation(operationToken *lex.Token, left, right core.Filter) (core.Filter, error) {
	operation := string(operationToken.Lexeme)
	if operationToken.Type != typeCondOperation {
		return nil, errors.Errorf("Unexpected token type for logic operation '%s'", operation)
	}
	switch operation {
	case opAnd:
		return &And{
			Left:  left,
			Right: right,
		}, nil
	case opOr:
		return &Or{
			Left:  left,
			Right: right,
		}, nil
	default:
		return nil, errors.Errorf("Unknown logic operation '%s'", operation)
	}
}
