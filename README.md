[![GoDoc](https://godoc.org/github.com/Zaba505/gqlc?status.svg)](https://godoc.org/github.com/Zaba505/gqlc)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zaba505/gqlc)](https://goreportcard.com/report/github.com/Zaba505/gqlc)
[![Build Status](https://travis-ci.org/Zaba505/gqlc.svg?branch=master)](https://travis-ci.org/Zaba505/gqlc)
[![codecov](https://codecov.io/gh/Zaba505/gqlc/branch/master/graph/badge.svg)](https://codecov.io/gh/Zaba505/gqlc)

# GraphQL IDL Compiler

`gqlc` is a compiler for the GraphQL IDL, as defined by the [GraphQL spec](http://facebook.github.io/graphql).
Current spec implementation: Current Working Draft i.e. >June2018

# Table of Contents

- [Usage](#usage)
    * [Support Languages](*supported-languages)
- [Design](#design)
    * [IDL packages](#idl-pacakges)
    * [Code generation](#code-generation-and-cli)
- [WIP](#wip)

## Usage
To use `gqlc`, there are two options: the `gqlc` cli tool or writing your own
cli. In order to use the `gqlc` cli tool you must either download a pre-built
[binary]() or if you are familiar using the Go toolchain: `go get github.com/Zaba505/gqlc/cmd/gqlc`

Example:
```text
gqlc -I . --dart_out ./dartapi
        \ --doc_out ./docs
        \ --go_out ./goapi
        \ --js_out ./jsapi
        \ api.gql
```

### Supported Languages
The currently supported languages by gqlc for generation are:

* [Dart](https://dartlang.org)
* [Documentation](https://commonmark.org)
* [Go](https://golang.org)
* [Javascript](https://javascript.com)

There will most likely be more to come. Feel free to submit an issue to
discuss supporting your language of choice.

## Design

The overall design of the compiler is heavily influenced by [Google's Protocol Buffer](https://github.com/protocolbuffers/protobuf) compiler.

### IDL Pacakges

Overall structure and "connected-ness" is heavily influenced by Go's [go](https://golang.org/pkg/go) package for working with Go source files.
The lexer and parser are implemented following the [text/template/parse](https://golang.org/pkg/text/template/parse) package
and Rob Pike's talk on ["Lexical Scanning in Go"](https://talks.golang.org/2011/lex.slide).

### Code Generation and CLI

The code generation and CLI designs were both pulled from Google's Protocol Buffer compiler, in order
to allow for extensibility and ease of maintainability.

## WIP
This is all current work desired to be completed in order to release v1.

- [ ] cmd/gqlc
    - [ ] type checking
    - [x] generator options flag
    - [ ] support plugins
- [ ] compiler
    - [ ] Dart generator
    - [ ] Documentation generator
    - [ ] Go generator
    - [ ] Javascript generator
- [x] graphql
    - [x] ast
    - [x] lexer
    - [x] parser
    - [x] token