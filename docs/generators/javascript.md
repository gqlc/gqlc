# JavaScript

Generates all the types needed for implementing a GraphQL service in [JavaScript](https://javascript.com).
All types are generated to work with [github.com/graphql/graphql-js](https://github.com/graphql/graphql-js).

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
```javascript
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

## Options

### descriptions
  - **Type:** `Boolean`

  - **Default:** `false`

  Keep descriptions in generated output.

### module
  - **Type:** `Enum`

  - **Default:** `COMMONJS`

  - **Values:**
    - `COMMONJS`
    - `ES6`

  Specify the Javascript module import style.

### useFlow
  - **Type:** `Boolean`

  - **Default:** `false`

  Add [flow](https://flow.org) comment.
