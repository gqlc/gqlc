[![GoDoc](https://godoc.org/github.com/gqlc/gqlc?status.svg)](https://godoc.org/github.com/gqlc/gqlc)
[![Go Report Card](https://goreportcard.com/badge/github.com/gqlc/gqlc)](https://goreportcard.com/report/github.com/gqlc/gqlc)
[![Build Status](https://travis-ci.org/gqlc/gqlc.svg?branch=master)](https://travis-ci.org/gqlc/gqlc)
[![codecov](https://codecov.io/gh/gqlc/gqlc/branch/master/graph/badge.svg)](https://codecov.io/gh/gqlc/gqlc)

# GraphQL IDL Compiler

`gqlc` is a compiler for the GraphQL IDL, as defined by the [GraphQL spec](http://facebook.github.io/graphql).
Current spec implementation: [Current Working Draft](https://graphql.github.io/graphql-spec/draft/)

# Table of Contents

- [Getting Started](#getting-started)
  * [Installing](#installing)
  * [Compiling Your First Schema](#compiling-your-first-schema)
- [Support Languages](#supported-languages)
- [Design](#design)
    * [IDL packages](#idl-pacakges)
    * [Code generation](#code-generation-and-cli)
- [Contributing](#contributing)
    * [Getting Started](#getting-started)
        - [Guidelines](#guidelines)
        - [Code Generators](#code-generators)
- [WIP](#wip)

## Getting Started
This section gives a brief intro to using gqlc.

### Installing
You can either `git clone` this repo and build from source or download one of the prebuilt [releases](https://github.com/gqlc/gqlc/releases).

### Compiling Your First Schema
To begin, lets use an abbreviated version of the schema used in the examples at [graphql.org](https://graphql.org/learn/schema/):

```graphql
schema {
  query: Query,
  mutation: Mutation
}

type Query {
  "hero returns a character in an episode"
  hero(episode: Episode): Character
}

type Mutation {
  """
  addCharacter adds a new Character given their name and the episodes they appeared in.
  """
  addCharacter(name: String!, episodes: [Episode!]!): Character
}

enum Episode {
  NEWHOPE
  EMPIRE
  JEDI
}

type Character {
  name: String!
  appearsIn: [Episode]!
}
```

Now, that we have the schema for our GraphQL service it's time to start
implementing it. Typically, when implementing a GraphQL service you're thinking
in terms of the IDL, but not writing in it; instead, you're writing in whatever
language you have chosen to implement your service in. This is where `gqlc`
comes in handy, by providing you with a tool that can "compile", or translate,
your IDL definitions into source code definitions. To accomplish this, simply
type the following into your shell:

```bash
gqlc --js_out ./js_service
   \ --go_out ./go_service
   \ --doc_out ./docs
   \ schema.gql
```

`gqlc` will then generate three directories:
- *js_service*: The js generator generates Javascript types for the schema.
- *go_service*: The go generator generates Go types for the schema.
- *docs*: The doc generator generates Commonmark documentation.

## Supported Languages
The currently supported languages by gqlc for generation are:

* [Documentation](https://commonmark.org) ([example](https://github.com/gqlc/doc#example))
* [Go](https://golang.org)                ([example](https://github.com/gqlc/golang#example))
* [Javascript](https://javascript.com)    ([example](https://github.com/gqlc/js#example))

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
