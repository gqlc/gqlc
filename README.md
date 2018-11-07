# GraphQL Schema Language Compiler

`gqlc` is a compiler for the "GraphQL Schema Language", as defined by the [GraphQL spec](http://facebook.github.io/graphql).
Current spec implementation: Current Working Draft i.e. >June2018

## Design

The overall design of the compiler is heavily influenced by [Google's Protocol Buffer](https://github.com/protocolbuffers/protobuf) compiler.

#### IDL Pacakges

Overall structure and "connected-ness" is heavily influenced by Go's [go](https://golang.org/pkg/go) package for working with Go source files.
The lexer and parser are implemented following the [text/template/parse](https://golang.org/pkg/text/template/parse) package
and Rob Pike's talk on "Lexical Scanning in Go", which the slides for can be found [here](https://talks.golang.org/2011/lex.slide).

#### Code Generation and CLI

The code generation and CLI designs were both pulled from Google's Protocol Buffer compiler, in order
to allow for extensibility and ease of maintainability.