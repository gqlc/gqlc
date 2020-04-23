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
