[![GoDoc](https://godoc.org/github.com/gqlc/gqlc?status.svg)](https://godoc.org/github.com/gqlc/gqlc)
[![Go Report Card](https://goreportcard.com/badge/github.com/gqlc/gqlc)](https://goreportcard.com/report/github.com/gqlc/gqlc)
[![Build Status](https://travis-ci.org/gqlc/gqlc.svg?branch=master)](https://travis-ci.org/gqlc/gqlc)
[![codecov](https://codecov.io/gh/gqlc/gqlc/branch/master/graph/badge.svg)](https://codecov.io/gh/gqlc/gqlc)

# GraphQL IDL Compiler

`gqlc` is a compiler for the GraphQL IDL, as defined by the [GraphQL spec](http://facebook.github.io/graphql).
Current spec implementation: [Current Working Draft](https://graphql.github.io/graphql-spec/draft/)

# Table of Contents

- [Installing](#installing)
- [Usage](#usage)
    * [Support Languages](*supported-languages)
- [Design](#design)
    * [IDL packages](#idl-pacakges)
    * [Code generation](#code-generation-and-cli)
- [Contributing](#contributing)
    * [Getting Started](#getting-started)
        - [Guidelines](#guidelines)
        - [Code Generators](#code-generators)
- [WIP](#wip)

## Installing
You can either `git clone` this repo and build from source or download one of the prebuilt [releases](https://github.com/gqlc/gqlc/releases).

## Usage
To use `gqlc`, all that's needed is a GraphQL Document (only type system defs) and a directory to output generated code to.

Example:
```text
gqlc -I . --doc_out ./docs
        \ --go_out ./goapi
        \ --js_out ./jsapi
        \ api.gql
```

### Supported Languages
The currently supported languages by gqlc for generation are:

* [Documentation](https://commonmark.org)
* [Go](https://golang.org)
* [Javascript](https://javascript.com)

*Note*: There will most likely be more to come. Check out the [Code Generators](#code-generators) section for more on this.

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
When making any sort of contribution remember to follow the [Contribution guidelines](https://github.com/gqlc/gqlc/blob/master/CONTRIBUTING.md).

#### Code Generators
Not every language can be supported directly, so please first create an issue and discuss adding support for your language there.
If the community shows enough consensus that your language should be directly supported, then a @gqlc team member will initialize
the repository for it and work can commence on implementing it.

If your desired language doesn't show enough support from the community to deem direct support in gqlc, then implementing a plugin
is highly encouraged. Check out [compiler/plugin](https://github.com/gqlc/compiler/tree/master/plugin) for more information on how
plugins are expected to behave when interacting with gqlc.

## WIP
This is all current work desired to be completed in order to release v1.

- [x] gqlc ([cmd](https://github.com/gqlc/gqlc/tree/master/cmd))
    - [x] generator options flag
    - [x] Plugin generator ([plugin](https://github.com/gqlc/gqlc/tree/master/cmd/plugin))
- [x] Documentation generator ([doc](https://github.com/gqlc/doc))
- [x] Go generator ([golang](https://github.com/gqlc/golang))
- [x] Javascript generator ([js](https://github.com/gqlc/js))
- [x] [compiler](https://github.com/gqlc/compiler)
    - [x] type checking
    - [x] import resolution
- [x] [graphql](https://github.com/gqlc/graphql)
    - [x] ast
    - [x] lexer
    - [x] parser
    - [x] token