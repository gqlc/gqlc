[![GoDoc](https://godoc.org/github.com/gqlc/golang?status.svg)](https://godoc.org/github.com/gqlc/golang)
[![Go Report Card](https://goreportcard.com/badge/github.com/gqlc/golang)](https://goreportcard.com/report/github.com/gqlc/golang)
[![Build Status](https://travis-ci.org/gqlc/golang.svg?branch=master)](https://travis-ci.org/gqlc/golang)
[![codecov](https://codecov.io/gh/gqlc/golang/branch/master/graph/badge.svg)](https://codecov.io/gh/gqlc/golang)

# Go Generator

This generates Go from a GraphQL Document.

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
```go
package main
import "github.com/graphql-go/graphql"

var Schema graphql.Schema

var QueryType = graphql.NewObject(graphql.ObjectConfig{
 	Name: "Query",
	Fields: graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil }, // TODO
		},
	},
	Description: "Query represents the queries this example provides.",
})

func init() {
	var err error
	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: QueryType,
	})
	if err != nil {
		panic(err)
	}
}
```