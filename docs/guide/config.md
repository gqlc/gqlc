# Config

`gqlc` and the generators provide options to the user for customizing the
generated output.

## CLI Options

One of the simplest to set an option, primarily for generators, is by setting
them when using `gqlc`. Setting a generator option at the command line can be
done by the following:
```bash
gqlc --doc_out=html:. api.gql
# or
gqlc --doc_out . --doc_opt html api.gql
```

## Directive Options

Generator options can also be set in your `.gql` file, as well as, global `gqlc`
options. For example:

```graphql
@doc(options: {
  # set the title of the documentation
  title: "Your Service Documentation",
  # render documentation to html, as well as, markdown
  html: true
})

schema {
  query: Query
}

type Query {
  echo(msg: String): String! @resolver(name: echo)
}
```

The `resolver` directive is a `gqlc` specific directive that is used for telling
the generators that the resolver field should be set to the given function name.
Its main purpose is to seperate resolvers from generated code.
