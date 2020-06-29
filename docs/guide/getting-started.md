# Getting Started

## Creating a Schema

The usage of `gqlc` doesn't actually begin with `gqlc` itself but instead with
you, the developer, thinking about your services' API. For example, lets say
one day you're tasked with implementing a service that manages all information
related to the Star Wars franchise. You may begin by roughly outling your
schema as the following:

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
  "addCharacter adds a new Character given their name and the episodes they appeared in."
  addCharacter(name: String!, episodes: [Episode!]!): Character
}

"Episode represents the episodes of the Star Wars saga."
enum Episode {
  NEWHOPE
  EMPIRE
  JEDI
}

"Character represents a character in any episode of Star Wars."
type Character {
  name: String!
  appearsIn: [Episode]!
}
```

## Generating a Service

Once you think your API is complete (or complete enough), `gqlc` can now be used
to speed up your development time by generating the scaffolding of your service
based on the above GraphQL IDL. Let's say your language of choice is JavaScript,
then simply run the following:
```bash
gqlc --js_out . api.gql
```

The above command will generate a file, *api.js*, that contains the scaffolding
on which you implement the business logic for the service.

```javascript
var {
  GraphQLSchema,
  GraphQLObjectType,
  GraphQLEnumType,
  GraphQLList,
  GraphQLNonNull,
  GraphQLString
} = require('graphql');

var Schema = new GraphQLSchema({
  query: Query,
  mutation: Mutation
});

var EpisodeType = new GraphQLEnumType({
  name: 'Episode',
  values: {
    NEWHOPE: {
      value: 'NEWHOPE'
    },
    EMPIRE: {
      value: 'EMPIRE'
    },
    JEDI: {
      value: 'JEDI'
    }
  }
});

var QueryType = new GraphQLObjectType({
  name: 'Query',
  fields: {
    hero: {
      type: Character,
      args: {
        episode: {
          type: Episode
        }
      },
      resolve() { /* TODO */ }
    }
  }
});

var MutationType = new GraphQLObjectType({
  name: 'Mutation',
  fields: {
    addCharacter: {
      type: Character,
      args: {
        name: {
          type: new GraphQLNonNull(GraphQLString)
        },
        episodes: {
          type: new GraphQLNonNull(new GraphQLList(new GraphQLNonNull(Episode)))
        }
      },
      resolve() { /* TODO */ }
    }
  }
});

var CharacterType = new GraphQLObjectType({
  name: 'Character',
  fields: {
    name: {
      type: new GraphQLNonNull(GraphQLString),
      resolve() { /* TODO */ }
    },
    appearsIn: {
      type: new GraphQLNonNull(new GraphQLList(Episode)),
      resolve() { /* TODO */ }
    }
  }
});
```

All that's left is to fill in the resolvers. Then, it's on its way to deployment!
