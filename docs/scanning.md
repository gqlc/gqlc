# Lexical Scanning and Parsing

The GraphQL Schema Language lexer and parser implemented in `gqlc` are
designed based upon the implementation in the Golang standard library package
`text/template/parse` and Rob Pike's talk, [Lexical Scanning in Go](https://talks.golang.org/2011/lex.slide#1).

### Lexer

The goal of the lexer is to simply scan tokens and emit them to the parser.

### Parser

The goal of the parser is to analyze incoming tokens and assemble a parse
tree representing the provided text.