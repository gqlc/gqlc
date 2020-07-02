# Go

Generates all the types needed for implementing a GraphQL service in [Go](https://golang.org).
All types are generated to work with [github.com/graphql-go/graphql](https://github.com/graphql-go/graphql).

## Output

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
```

## Options

### descriptions
  - **Type:** `Boolean`

  - **Default:** `false`

  Keep descriptions in generated output.

### package
  - **Type:** `String`

  - **Default:** `main`

  Package name to be used when generating.
