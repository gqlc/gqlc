# Remote Service as a Source

`gqlc` supports using a running GraphQL service/endpoint as a source. This can be
used simply for documenting the running service like:
```bash
gqlc --doc_out . https://api.example.com/graphql
```

Or more extremely to migrate a service implementation from language to another like:
```bash
# Assume it's originally implemented in JavaScript
# and your boss wants it migrated to Go.
gqlc --go_out . https://api.example.com/graphql
```

`gqlc` also supports Apollos' [GraphQL over Websocket](https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md) protocol.

```bash
gqlc --doc_out . wss://api.example.com/graphql
```
