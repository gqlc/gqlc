[![GoDoc](https://godoc.org/github.com/gqlc/js?status.svg)](https://godoc.org/github.com/gqlc/js)
[![Go Report Card](https://goreportcard.com/badge/github.com/gqlc/js)](https://goreportcard.com/report/github.com/gqlc/js)
[![Build Status](https://travis-ci.org/gqlc/js.svg?branch=master)](https://travis-ci.org/gqlc/js)
[![codecov](https://codecov.io/gh/gqlc/js/branch/master/graph/badge.svg)](https://codecov.io/gh/gqlc/js)

# JavaScript Generator

This generates Javascript from a GraphQL Document.

## Example

Input:
```graphql
schema {
	query: Query
}

"Query represents the queries this example provides."
type Query {
	hello: String
}
```

Output:
```js
// @flow

var {
  GraphQLSchema,
  GraphQLObjectType,
  GraphQLString
} = require('graphql');

var Schema = new GraphQLSchema({
  query: Query
});

var QueryType = new GraphQLObjectType({
  name: 'Query',
  fields: {
    hello: {
      type: GraphQLString,
      resolve() { /* TODO */ }
    }
  }
});
```