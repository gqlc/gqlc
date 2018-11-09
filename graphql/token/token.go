// Package token defines constants representing the lexical tokens of the GraphQL IDL and basic operations on tokens (printing, predicates).
package token

import "strconv"

// Token is the set of lexical tokens of the GraphQL Schema Language
type Token int

// A list of tokens
const (
	// Special Tokens
	ERR Token = iota
	EOF
	COMMENT
	DESCRIPTION

	litBeg
	IDENT  // query
	STRING // "abc" or """abc"""
	INT    // 123
	FLOAT  // 123.45
	litEnd

	opBeg
	AND // &
	OR  // |
	NOT // !
	AT  // @
	VAR // $

	ASSIGN // =
	LPAREN // (
	LBRACK // [
	LBRACE // {
	COMMA  // ,

	PERIOD // .
	RPAREN // )
	RBRACK // ]
	RBRACE // }
	COLON  // :
	opEnd

	keyBeg
	PACKAGE
	IMPORT
	SCHEMA
	TYPE

	SCALAR
	ENUM
	INTERFACE
	IMPLEMENTS
	UNION

	INPUT
	EXTEND
	DIRECTIVE
	ON
	keyEnd
)

var tokens = [...]string{
	ERR: "ERROR",

	EOF:         "EOF",
	DESCRIPTION: "DESCRIPTION",

	IDENT:  "IDENT",
	STRING: "STRING",

	AND:    "&",
	OR:     "|",
	NOT:    "!",
	AT:     "@",
	ASSIGN: "=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",
	COMMA:  ",",
	PERIOD: ".",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",
	COLON:  ":",

	PACKAGE: "package",
	IMPORT:  "import",
	SCHEMA:  "schema",
	TYPE:    "type",

	SCALAR:     "scalar",
	ENUM:       "enum",
	INTERFACE:  "interface",
	IMPLEMENTS: "implements",
	UNION:      "union",

	INPUT:     "input",
	EXTEND:    "extend",
	DIRECTIVE: "directive",
	ON:        "on",
}

func (tok Token) String() string {
	s := ""
	if 0 <= tok && tok < Token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyBeg + 1; i < keyEnd; i++ {
		keywords[tokens[i]] = i
	}
}

func Lookup(ident string) Token {
	if tok, isKeyword := keywords[ident]; isKeyword {
		return tok
	}
	return IDENT
}

// Predicates

// IsLiteral returns true for tokens corresponding to identifiers
// and basic type literals; it returns false otherwise.
//
func (tok Token) IsLiteral() bool { return litBeg < tok && tok < litEnd }

// IsOperator returns true for tokens corresponding to operators and
// delimiters; it returns false otherwise.
//
func (tok Token) IsOperator() bool { return opBeg < tok && tok < opEnd }

// IsKeyword returns true for tokens corresponding to keywords;
// it returns false otherwise.
//
func (tok Token) IsKeyword() bool { return keyBeg < tok && tok < keyEnd }
