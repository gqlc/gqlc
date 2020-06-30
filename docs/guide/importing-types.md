# Importing Types

**Note:** This is not a GraphQL IDL spec compliant feature.

`gqlc` supports importing types from a given source. To remain as close to spec
compliant as possible imports do not inhabit any sort of namepsace i.e. there is
only one global namespace. This will most likely change in the future given recent
work on [graphq-spec#163](https://github.com/graphql/graphql-spec/issues/163).

To import types its as simple as using the following directive:
```graphql
@directive import(paths: [String!]!) on DOCUMENT
```

For example:
```graphql
@import(paths: [
  "service-a.gql"
  "service-b.gql"
])

schema {
  query: Query
}

type Query {
  a: TypeFromServiceA
  b: TypeFromServiceB
}
```
