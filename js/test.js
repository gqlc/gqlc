// @flow

var {
  GraphQLSchema,
  GraphQLScalarType,
  GraphQLObjectType,
  GraphQLInterfaceType,
  GraphQLUnionType,
  GraphQLEnumType,
  GraphQLInputObjectType,
  GraphQLDirectiveType,
  GraphQLList,
  GraphQLNonNull,
  GraphQLInt,
  GraphQLFloat,
  GraphQLString,
  GraphQLBoolean,
  GraphQLID
} = require('graphql');

var Schema = new GraphQLSchema({
  query: Query
});

var VersionType = new GraphQLScalarType({
  name: 'Version',
  description: 'Version represents an API version.',
  serialize(value) { /* TODO */ }
});

var EchoType = new GraphQLObjectType({
  name: 'Echo',
  fields: {
    msg: {
      type: new GraphQLNonNull(GraphQLString),
      resolve() { /* TODO */ },
      description: 'msg contains the provided message.'
    }
  },
  description: 'Echo represents an echo message.'
});

var QueryType = new GraphQLObjectType({
  name: 'Query',
  fields: {
    version: {
      type: Version,
      resolve() { /* TODO */ },
      description: 'version returns the current API version.'
    },
    echo: {
      type: Echo,
      args: {
        text: {
          type: new GraphQLNonNull(GraphQLString)
        }
      },
      resolve() { /* TODO */ },
      description: 'echo echos a message.'
    },
    search: {
      type: Result,
      args: {
        text: {
          type: GraphQLString,
          description: 'text is a single text input to use for searching.'
        },
        terms: {
          type: new GraphQLList(GraphQLString),
          description: 'terms represent term based querying.'
        }
      },
      resolve() { /* TODO */ },
      description: 'search performs a search over some data set.'
    }
  },
  description: 'Query represents valid queries.'
});

var ResultType = new GraphQLObjectType({
  name: 'Result',
  interfaces: [ Connection ],
  fields: {
    total: {
      type: GraphQLInt,
      resolve() { /* TODO */ },
      description: 'total yields the total number of search results.'
    },
    edges: {
      type: new GraphQLList(Node),
      resolve() { /* TODO */ },
      description: 'edges contains the search results.'
    },
    hasNextPage: {
      type: GraphQLBoolean,
      resolve() { /* TODO */ },
      description: 'hasNextPage tells if there are more search results.'
    }
  },
  description: 'Result represents a search result.'
});

var ConnectionType = new GraphQLInterfaceType({
  name: 'Connection',
  fields: {
    total: {
      type: GraphQLInt,
      description: 'total returns the total number of edges.'
    },
    edges: {
      type: new GraphQLList(Node),
      description: 'edges contains the current page of edges.'
    },
    hasNextPage: {
      type: GraphQLBoolean,
      description: 'hasNextPage tells if there exists more edges.'
    }
  },
  description: 'Connection represents a set of edges, which are meant to be paginated.'
});

var NodeType = new GraphQLInterfaceType({
  name: 'Node',
  fields: {
    id: {
      type: new GraphQLNonNull(GraphQLID),
      description: 'id uniquely identifies the node.'
    }
  },
  description: 'Node represents a node.'
});

var SearchResultType = new GraphQLUnionType({
  name: 'SearchResult',
  types: [
    Echo,
    Result
  ],
  resolveType(value) { /* TODO */ },
  description: 'SearchResult is a test union type'
});

var DirectionType = new GraphQLEnumType({
  name: 'Direction',
  description: 'Direction represents a cardinal direction.',
  values: {
    NORTH: {
      value: 'NORTH',
      description: 'EnumValue description'
    },
    EAST: {
      value: 'EAST'
    },
    SOUTH: {
      value: 'SOUTH'
    },
    WEST: {
      value: 'WEST',
      description: 'EnumValue Description and Directives.'
    }
  }
});

var PointType = new GraphQLInputObjectType({
  name: 'Point',
  fields: {
    x: {
      type: new GraphQLNonNull(GraphQLFloat)
    },
    y: {
      type: new GraphQLNonNull(GraphQLFloat)
    }
  },
  description: 'Point represents a 2-D geo point.'
});

var deprecateType = new GraphQLDirectiveType({
  name: 'deprecate',
  description: 'deprecate signifies a type deprecation from the api.',
  locations: [
    DirectiveLocation.SCHEMA,
    DirectiveLocation.FIELD
  ],
  args: {
    msg: {
      type: GraphQLString,
      description: 'Arg description.'
    }
  }
});
