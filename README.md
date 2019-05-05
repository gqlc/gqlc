[![GoDoc](https://godoc.org/github.com/gqlc/gqlc?status.svg)](https://godoc.org/github.com/gqlc/gqlc)
[![Go Report Card](https://goreportcard.com/badge/github.com/gqlc/gqlc)](https://goreportcard.com/report/github.com/gqlc/gqlc)
[![Build Status](https://travis-ci.org/gqlc/gqlc.svg?branch=master)](https://travis-ci.org/gqlc/gqlc)
[![codecov](https://codecov.io/gh/gqlc/gqlc/branch/master/graph/badge.svg)](https://codecov.io/gh/gqlc/gqlc)

# GraphQL IDL Compiler

`gqlc` is a compiler for the GraphQL IDL, as defined by the [GraphQL spec](http://facebook.github.io/graphql).
Current spec implementation: Current Working Draft i.e. >June2018

# Table of Contents

- [Usage](#usage)
    * [Support Languages](*supported-languages)
- [Design](#design)
    * [IDL packages](#idl-pacakges)
    * [Code generation](#code-generation-and-cli)
- [Contributing](#contributing)
    * [Getting Started](#getting-started)
        - [Guidelines](#guidelines)
        - [Code Generators](#code-generators)
    * [WIP](#wip)

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
and Rob Pike's talk on ["Lexical Scanning in Go"](https://talks.golang.org/2011/lex.slide). The source for the lexer and parser can be
found here: [graphql](https://github.com/gqlc/graphql)

### Code Generation and CLI

The code generation and CLI designs were both pulled from Google's Protocol Buffer compiler, in order
to allow for extensibility and ease of maintainability. The source for internal compiler details and cli
interfaces can be found here: [compiler](https://github.com/gqlc/compiler)

## Contributing

### Getting Started

Thank you for wanting to help keep this project awesome!

Before diving straight into the [WIP](#wip) list or issues, here are a few things to help your contribution be accepted:

#### Guidelines
When making any sort of contribution remember to follow the [Contribution guidelines]().

#### Code Generators
Not every language can be supported, so please first create an issue and discuss adding support for your language there.
If the community shows enough consensus that your language should be directly supported then a PR can be submitted/accepted. 

If you desire to contribute a code generator, a @gqlc team member will init a repo in the @gqlc org
for it. Once the generator is complete and tested, it can be registered with the CLI.

### WIP
This is all current work desired to be completed in order to release v1.

- [x] cmd/gqlc
    - [x] generator options flag
    - [x] support plugins
- [ ] compiler
    - [ ] type checking
    - [x] import resolution
    - [ ] Dart generator
    - [x] Documentation generator
    - [ ] Go generator
    - [ ] Javascript generator
- [x] graphql
    - [x] ast
    - [x] lexer
    - [x] parser
    - [x] token