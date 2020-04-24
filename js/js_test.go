package js

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
)

var (
	update = flag.Bool("update", false, "Update expected output file")

	// Flags are used here to allow for the input/output files to be changed during dev
	// One use case of changing the files is to examine how Generate scales through the benchmark
	//
	gqlFileName = flag.String("gqlFile", "test.gql", "Specify a .gql file to use a input for testing.")
	exDocName   = flag.String("expectedFile", "test.js", "Specify a file which is the expected generator output from the given .gql file.")

	testDoc *ast.Document
	exDoc   io.Reader
)

func TestMain(m *testing.M) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Parse flags
	flag.Parse()

	// Assume the input file is in the current working directory
	if !filepath.IsAbs(*gqlFileName) {
		*gqlFileName = filepath.Join(wd, *gqlFileName)
	}
	f, err := os.Open(*gqlFileName)
	if err != nil {
		panic(err)
	}

	// Assume the output file is in the current working directory
	if !filepath.IsAbs(*exDocName) {
		*exDocName = filepath.Join(wd, *exDocName)
	}
	exDoc, err = os.Open(*exDocName)
	if err != nil {
		panic(err)
	}

	testDoc, err = parser.ParseDoc(token.NewDocSet(), "test", f, 0)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestUpdate(t *testing.T) {
	if !*update {
		t.Skipf("not updating expected js output file: %s", *exDocName)
		return
	}
	t.Logf("updating expected js output file: %s", *exDocName)

	f, err := os.OpenFile(*exDocName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		t.Error(err)
		return
	}

	g := new(Generator)
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: f})
	err = g.Generate(ctx, testDoc, map[string]interface{}{"descriptions": true})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestImports(t *testing.T) {
	g := &Generator{}

	t.Run("CommonJS", func(subT *testing.T) {

		subT.Run("Single", func(triT *testing.T) {
			imports := [][]byte{schemaImport}

			var b bytes.Buffer
			_, err := g.writeImports(&b, &Options{Module: "COMMONJS", imports: imports})
			if err != nil {
				triT.Error(err)
				return
			}

			ex := []byte("var { GraphQLSchema } = require('graphql');\n\n")
			gen.CompareBytes(triT, ex, b.Bytes())
		})

		subT.Run("Multiple", func(triT *testing.T) {
			imports := [][]byte{schemaImport, scalarImport}

			var b bytes.Buffer
			_, err := g.writeImports(&b, &Options{Module: "COMMONJS", imports: imports})
			if err != nil {
				triT.Error(err)
				return
			}

			ex := []byte("var {\n  GraphQLSchema,\n  GraphQLScalarType\n} = require('graphql');\n\n")
			gen.CompareBytes(triT, ex, b.Bytes())
		})
	})

	t.Run("ES6", func(subT *testing.T) {

		subT.Run("Single", func(triT *testing.T) {
			imports := [][]byte{schemaImport}

			var b bytes.Buffer
			_, err := g.writeImports(&b, &Options{Module: "ES6", imports: imports})
			if err != nil {
				triT.Error(err)
				return
			}

			ex := []byte("import { GraphQLSchema } from 'graphql';\n\n")
			gen.CompareBytes(triT, ex, b.Bytes())
		})

		subT.Run("Multiple", func(triT *testing.T) {
			imports := [][]byte{schemaImport, scalarImport}

			var b bytes.Buffer
			_, err := g.writeImports(&b, &Options{Module: "ES6", imports: imports})
			if err != nil {
				triT.Error(err)
				return
			}

			ex := []byte("import {\n  GraphQLSchema,\n  GraphQLScalarType\n} from 'graphql';\n\n")
			gen.CompareBytes(triT, ex, b.Bytes())
		})
	})
}

func TestSchema(t *testing.T) {
	g := &Generator{}

	t.Run("WithoutMutation", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{
			Type: &ast.TypeSpec_Schema{
				Schema: &ast.SchemaType{
					RootOps: &ast.FieldList{List: []*ast.Field{
						{Name: &ast.Ident{Name: "query"}, Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Query"}}},
					}},
				},
			},
		}

		g.generateSchema(&Options{Module: "COMMONJS", declStr: commonJSDecl}, ts)

		ex := []byte(`var Schema = new GraphQLSchema({
  query: Query
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

	t.Run("WithMutation", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{
			Type: &ast.TypeSpec_Schema{
				Schema: &ast.SchemaType{
					RootOps: &ast.FieldList{List: []*ast.Field{
						{Name: &ast.Ident{Name: "query"}, Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Query"}}},
						{Name: &ast.Ident{Name: "mutation"}, Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Mutation"}}},
					}},
				},
			},
		}

		g.generateSchema(&Options{Module: "COMMONJS", declStr: commonJSDecl}, ts)

		ex := []byte(`var Schema = new GraphQLSchema({
  query: Query,
  mutation: Mutation
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

}

func TestScalar(t *testing.T) {
	g := &Generator{}

	ts := &ast.TypeSpec{
		Name: &ast.Ident{Name: "Test"},
	}

	g.generateScalar(nil, "Test", false, nil, ts)

	ex := []byte(`GraphQLScalarType({
  name: 'Test',
  serialize(value) { /* TODO */ }
});
`)

	gen.CompareBytes(t, ex, g.Bytes())
}

func TestObject(t *testing.T) {
	g := &Generator{}

	t.Run("JustFields", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Object{
			Object: &ast.ObjectType{
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Name: &ast.Ident{Name: "one"},
							Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Int"}},
						},
						{
							Name: &ast.Ident{Name: "str"},
							Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
						},
						{
							Name: &ast.Ident{Name: "list"},
							Type: &ast.Field_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
						},
						{
							Name: &ast.Ident{Name: "withDefaultVal"},
							Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
							Args: &ast.InputValueList{List: []*ast.InputValue{{
								Name:    &ast.Ident{Name: "str"},
								Type:    &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
								Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{Value: `"hello"`}},
							}}},
						},
					},
				},
			},
		}}

		g.generateObject(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLObjectType({
  name: 'Test',
  fields: {
    one: {
      type: GraphQLInt,
      resolve() { /* TODO */ }
    },
    str: {
      type: GraphQLString,
      resolve() { /* TODO */ }
    },
    list: {
      type: new GraphQLList(Test),
      resolve() { /* TODO */ }
    },
    withDefaultVal: {
      type: GraphQLString,
      args: {
        str: {
          type: GraphQLString,
          defaultValue: "hello"
        }
      },
      resolve() { /* TODO */ }
    }
  }
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

	t.Run("WithInterfaces", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Object{
			Object: &ast.ObjectType{
				Interfaces: []*ast.Ident{{Name: "A"}, {Name: "B"}},
				Fields: &ast.FieldList{
					List: []*ast.Field{
						{
							Name: &ast.Ident{Name: "one"},
							Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Int"}},
						},
						{
							Name: &ast.Ident{Name: "str"},
							Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
						},
						{
							Name: &ast.Ident{Name: "list"},
							Type: &ast.Field_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
						},
					},
				},
			},
		}}

		g.generateObject(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLObjectType({
  name: 'Test',
  interfaces: [
    A,
    B
  ],
  fields: {
    one: {
      type: GraphQLInt,
      resolve() { /* TODO */ }
    },
    str: {
      type: GraphQLString,
      resolve() { /* TODO */ }
    },
    list: {
      type: new GraphQLList(Test),
      resolve() { /* TODO */ }
    }
  }
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})
}

func TestInterface(t *testing.T) {
	g := &Generator{}

	ts := &ast.TypeSpec{Type: &ast.TypeSpec_Interface{
		Interface: &ast.InterfaceType{
			Fields: &ast.FieldList{
				List: []*ast.Field{
					{
						Name: &ast.Ident{Name: "one"},
						Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Int"}},
					},
					{
						Name: &ast.Ident{Name: "str"},
						Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
					},
					{
						Name: &ast.Ident{Name: "list"},
						Type: &ast.Field_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
					},
				},
			},
		},
	}}

	g.generateInterface(new(uint16), "Test", false, nil, ts)

	ex := []byte(`GraphQLInterfaceType({
  name: 'Test',
  fields: {
    one: {
      type: GraphQLInt
    },
    str: {
      type: GraphQLString
    },
    list: {
      type: new GraphQLList(Test)
    }
  }
});
`)

	gen.CompareBytes(t, ex, g.Bytes())
}

func TestUnion(t *testing.T) {
	g := &Generator{}

	ts := &ast.TypeSpec{Type: &ast.TypeSpec_Union{
		Union: &ast.UnionType{
			Members: []*ast.Ident{{Name: "A"}, {Name: "B"}},
		},
	}}

	g.generateUnion(new(uint16), "Test", false, nil, ts)

	ex := []byte(`GraphQLUnionType({
  name: 'Test',
  types: [
    A,
    B
  ],
  resolveType(value) { /* TODO */ }
});
`)

	gen.CompareBytes(t, ex, g.Bytes())
}

func TestEnum(t *testing.T) {
	g := &Generator{}

	ts := &ast.TypeSpec{Type: &ast.TypeSpec_Enum{
		Enum: &ast.EnumType{
			Values: &ast.FieldList{
				List: []*ast.Field{
					{Name: &ast.Ident{Name: "A"}},
					{Name: &ast.Ident{Name: "B"}},
					{Name: &ast.Ident{Name: "C"}},
				},
			},
		},
	}}

	g.generateEnum(new(uint16), "Test", false, nil, ts)

	ex := []byte(`GraphQLEnumType({
  name: 'Test',
  values: {
    A: {
      value: 'A'
    },
    B: {
      value: 'B'
    },
    C: {
      value: 'C'
    }
  }
});
`)

	gen.CompareBytes(t, ex, g.Bytes())
}

func TestInput(t *testing.T) {
	g := &Generator{}

	t.Run("NoDefaults", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Input{
			Input: &ast.InputType{
				Fields: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "one"},
							Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
						},
						{
							Name: &ast.Ident{Name: "str"},
							Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
						},
						{
							Name: &ast.Ident{Name: "list"},
							Type: &ast.InputValue_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
						},
					},
				},
			},
		}}

		g.generateInput(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLInputObjectType({
  name: 'Test',
  fields: {
    one: {
      type: GraphQLInt
    },
    str: {
      type: GraphQLString
    },
    list: {
      type: new GraphQLList(Test)
    }
  }
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

	t.Run("WithDefaults", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Input{
			Input: &ast.InputType{
				Fields: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "one"},
							Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_INT,
								Value: "1",
							}},
						},
						{
							Name: &ast.Ident{Name: "str"},
							Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_STRING,
								Value: `"hello"`,
							}},
						},
						{
							Name: &ast.Ident{Name: "list"},
							Type: &ast.InputValue_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Int"}}}},
							Default: &ast.InputValue_CompositeLit{CompositeLit: &ast.CompositeLit{
								Value: &ast.CompositeLit_ListLit{
									ListLit: &ast.ListLit{
										List: &ast.ListLit_BasicList{
											BasicList: &ast.ListLit_Basic{
												Values: []*ast.BasicLit{
													{Kind: token.Token_INT, Value: "1"},
													{Kind: token.Token_INT, Value: "2"},
													{Kind: token.Token_INT, Value: "3"},
												},
											},
										},
									},
								},
							}},
						},
					},
				},
			},
		}}

		g.generateInput(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLInputObjectType({
  name: 'Test',
  fields: {
    one: {
      type: GraphQLInt,
      defaultValue: 1
    },
    str: {
      type: GraphQLString,
      defaultValue: 'hello'
    },
    list: {
      type: new GraphQLList(GraphQLInt),
      defaultValue: [1, 2, 3]
    }
  }
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})
}

func TestDirective(t *testing.T) {
	g := &Generator{}

	t.Run("NoArgs", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Directive{
			Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{
					{Loc: ast.DirectiveLocation_QUERY},
					{Loc: ast.DirectiveLocation_FIELD},
					{Loc: ast.DirectiveLocation_SCHEMA},
				},
			},
		}}

		g.generateDirective(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLDirectiveType({
  name: 'Test',
  locations: [
    DirectiveLocation.QUERY,
    DirectiveLocation.FIELD,
    DirectiveLocation.SCHEMA
  ]
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

	t.Run("WithArgs", func(subT *testing.T) {
		g.Lock()
		defer g.Unlock()
		g.Reset()

		ts := &ast.TypeSpec{Type: &ast.TypeSpec_Directive{
			Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{
					{Loc: ast.DirectiveLocation_QUERY},
					{Loc: ast.DirectiveLocation_FIELD},
					{Loc: ast.DirectiveLocation_SCHEMA},
				},
				Args: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name:    &ast.Ident{Name: "one"},
							Type:    &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{Value: "1"}},
						},
						{
							Name:    &ast.Ident{Name: "str"},
							Type:    &ast.InputValue_NonNull{NonNull: &ast.NonNull{Type: &ast.NonNull_Ident{Ident: &ast.Ident{Name: "String"}}}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{Kind: token.Token_STRING, Value: "\"hello\""}},
						},
						{
							Name: &ast.Ident{Name: "list"},
							Type: &ast.InputValue_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Int"}}}},
							Default: &ast.InputValue_CompositeLit{CompositeLit: &ast.CompositeLit{Value: &ast.CompositeLit_ListLit{
								ListLit: &ast.ListLit{
									List: &ast.ListLit_BasicList{BasicList: &ast.ListLit_Basic{Values: []*ast.BasicLit{
										{Value: "1"},
										{Value: "2"},
										{Value: "3"},
									}}},
								},
							}}},
						},
					},
				},
			},
		}}

		g.generateDirective(new(uint16), "Test", false, nil, ts)

		ex := []byte(`GraphQLDirectiveType({
  name: 'Test',
  locations: [
    DirectiveLocation.QUERY,
    DirectiveLocation.FIELD,
    DirectiveLocation.SCHEMA
  ],
  args: {
    one: {
      type: GraphQLInt,
      defaultValue: 1
    },
    str: {
      type: new GraphQLNonNull(GraphQLString),
      defaultValue: 'hello'
    },
    list: {
      type: new GraphQLList(GraphQLInt),
      defaultValue: [1, 2, 3]
    }
  }
});
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})
}

func TestGenerator_Generate(t *testing.T) {
	g := &Generator{}

	var b bytes.Buffer
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, map[string]interface{}{"descriptions": true})
	if err != nil {
		t.Error(err)
		return
	}

	ex, err := ioutil.ReadAll(exDoc)
	if err != nil {
		t.Error(err)
		return
	}

	gen.CompareBytes(t, ex, b.Bytes())
}

func BenchmarkGenerator_Generate(b *testing.B) {
	g := &Generator{}

	var buf bytes.Buffer
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &buf})

	for i := 0; i < b.N; i++ {
		buf.Reset()

		err := g.Generate(ctx, testDoc, nil)
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func ExampleGenerator_Generate() {
	g := new(Generator)

	gqlSrc := `schema {
	query: Query
}

"Query represents the queries this example provides."
type Query {
	hello: String
}`

	doc, err := parser.ParseDoc(token.NewDocSet(), "example", strings.NewReader(gqlSrc), 0)
	if err != nil {
		log.Fatal(err)
		return
	}

	var b bytes.Buffer
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b}) // Pass in an actual
	err = g.Generate(ctx, doc, map[string]interface{}{"module": "COMMONJS", "useFlow": true})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(b.String())

	// Output:
	// // @flow
	//
	// var {
	//   GraphQLSchema,
	//   GraphQLObjectType,
	//   GraphQLString
	// } = require('graphql');
	//
	// var Schema = new GraphQLSchema({
	//   query: Query
	// });
	//
	// var QueryType = new GraphQLObjectType({
	//   name: 'Query',
	//   fields: {
	//     hello: {
	//       type: GraphQLString,
	//       resolve() { /* TODO */ }
	//     }
	//   }
	// });
}
