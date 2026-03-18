package debug

import (
	"errors"
	"strconv"
	"strings"
)

// Operators with their priority
var operators = map[string]int{
	"(": 12, ")": 12,
	"*": 11, "/": 11, "%": 11,
	"+": 10, "-": 10,
	"<<": 9, ">>": 9,
	"<": 8, "<=": 8, ">": 8, ">=": 8,
	"=": 7, "!=": 7,
	"&":  6,
	"^":  5,
	"|":  4,
	"&&": 3,
	"||": 2,
}

var variables = map[string]bool{
	"A": true, "B": true, "C": true, "D": true, "E": true, "F": true, "H": true, "L": true, "I": true,
	"R": true, "SF": true, "NF": true, "PF": true, "VF": true, "XF": true, "YF": true, "ZF": true,
	"A'": true, "B'": true, "C'": true, "D'": true, "E'": true, "F'": true, "H'": true, "L'": true,
	"AF": true, "BC": true, "DE": true, "HL": true, "IX": true, "IY": true, "PC": true, "SP": true,
}

const (
	OTUnknown = iota
	OTValue
	OTVariable
	OTOperation
)

type Token struct {
	name string
	val  uint16
	ot   int
}

type Expression struct {
	infixExp string
	inStack  []Token
	outStack []Token
}

func NewExpression(infixExp string) *Expression {
	return &Expression{infixExp, make([]Token, 0), make([]Token, 0)}
}

func (e *Expression) Parse() error {
	e.infixExp = strings.ToUpper(strings.TrimSpace(e.infixExp))
	if e.infixExp == "" {
		return errors.New("no Expression")
	}
	ptr := 0
	for ptr < len(e.infixExp) {
		token, err := getNextToken(e.infixExp[ptr:])
		if err != nil {
			return err
		}
		err = validate(token)
		if err != nil {
			return err
		}
		err = e.parseToken(token)
		if err != nil {
			return err
		}
		ptr += len(token.name)
	}
	return nil
}

func (e *Expression) parseToken(token Token) error {
	return nil
}

func validate(token Token) error {
	switch token.ot {
	case OTValue:
		v, err := strconv.ParseUint(token.name, 0, 16)
		if err != nil {
			return err
		}
		token.val = uint16(v)
	case OTVariable:
		if !variables[token.name] {
			return errors.New("unknown variable")
		}
	case OTOperation:
		v, ok := operators[token.name]
		if !ok {
			return errors.New("unknown operation")
		}
		token.val = uint16(v)
	default:
		return errors.New("unknown token")
	}
	return nil
}

const operations = "*/%+_<=>!&^|"

func getNextToken(str string) (Token, error) {
	ptr := 0
	exp := ""
	ot := OTUnknown
	for ptr < len(str) {
		ch := str[ptr]
		if ch == ' ' {
			if ot == OTUnknown {
				ptr++
				continue
			} else {
				// end of token
				return Token{name: exp, ot: ot}, nil
			}
		}

		if (ch == 'X' || ch == 'O' || ch == 'B' || ch == 'H') && ot != OTValue {
			exp += string(ch)
			ptr++
			continue
		}

		if ch >= '0' && ch <= '9' {
			if len(exp) == 0 {
				ot = OTValue
			}
			exp += string(ch)
			ptr++
			continue
		}
		if strings.Contains(operations, string(ch)) {
			if len(exp) == 0 {
				ot = OTOperation
			}
			exp += string(ch)
			ptr++
			continue
		}
		return Token{name: exp, ot: ot}, errors.New("invalid token")
	}
	return Token{name: exp, ot: ot}, nil
}
