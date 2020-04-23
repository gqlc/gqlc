package golang

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
	exDocName   = flag.String("expectedFile", "test.gotxt", "Specify a file which is the expected generator output from the given .gql file.")

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
		t.Skipf("not updating expected go output file: %s", *exDocName)
		return
	}
	t.Logf("updating expected go output file: %s", *exDocName)

	f, err := os.OpenFile(*exDocName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		t.Error(err)
		return
	}

	g := new(Generator)
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: f})
	err = g.Generate(ctx, testDoc, `{"descriptions": true}`)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestScalar(t *testing.T) {
	g := &Generator{}

	ts := &ast.TypeSpec{
		Name: &ast.Ident{Name: "Test"},
	}

	g.generateScalar("Test", false, nil, ts)

	ex := []byte(`NewScalar(graphql.ScalarConfig{
	Name: "Test",
	Serialize: func(value interface{}) interface{} { return nil },
})
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

		g.generateObject("Test", false, nil, ts)

		ex := []byte(`NewObject(graphql.ObjectConfig{
	Name: "Test",
	Fields: graphql.Fields{
		"one": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
		"str": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
		"list": &graphql.Field{
			Type: graphql.NewList(TestType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
		"withDefaultVal": &graphql.Field{
			Type: graphql.String,
			Args: graphql.FieldConfigArgument{
				"str": &graphql.ArgumentConfig{
					Type: graphql.String,
					DefaultValue: "hello",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
	},
})
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

		g.generateObject("Test", false, nil, ts)

		ex := []byte(`NewObject(graphql.ObjectConfig{
	Name: "Test",
	Interfaces: []*graphql.Interface{
		AType,
		BType,
	},
	Fields: graphql.Fields{
		"one": &graphql.Field{
			Type: graphql.Int,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
		"str": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
		"list": &graphql.Field{
			Type: graphql.NewList(TestType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
		},
	},
})
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})

	t.Run("WithCustomResolver", func(subT *testing.T) {
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
							Directives: []*ast.DirectiveLit{{
								Name: "resolver",
								Args: &ast.CallExpr{Args: []*ast.Arg{{
									Name:  &ast.Ident{Name: "name"},
									Value: &ast.Arg_BasicLit{BasicLit: &ast.BasicLit{Value: `"customResolver"`}},
								}}},
							}},
						},
					},
				},
			},
		}}

		g.generateObject("Test", false, nil, ts)

		ex := []byte(`NewObject(graphql.ObjectConfig{
	Name: "Test",
	Fields: graphql.Fields{
		"one": &graphql.Field{
			Type: graphql.Int,
			Resolve: customResolver,
		},
	},
})
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

	g.generateInterface("Test", false, nil, ts)

	ex := []byte(`NewInterface(graphql.InterfaceConfig{
	Name: "Test",
	Fields: graphql.Fields{
		"one": &graphql.Field{
			Type: graphql.Int,
		},
		"str": &graphql.Field{
			Type: graphql.String,
		},
		"list": &graphql.Field{
			Type: graphql.NewList(TestType),
		},
	},
})
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

	g.generateUnion("Test", false, nil, ts)

	ex := []byte(`NewUnion(graphql.UnionConfig{
	Name: "Test",
	Types: []*graphql.Object{
		AType,
		BType,
	},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object { return nil },
})
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

	g.generateEnum("Test", false, nil, ts)

	ex := []byte(`NewEnum(graphql.EnumConfig{
	Name: "Test",
	Values: graphql.EnumValueConfigMap{
		"A": &graphql.EnumValueConfig{
			Value: "A",
		},
		"B": &graphql.EnumValueConfig{
			Value: "B",
		},
		"C": &graphql.EnumValueConfig{
			Value: "C",
		},
	},
})
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

		g.generateInput("Test", false, nil, ts)

		ex := []byte(`NewInputObject(graphql.InputObjectConfig{
	Name: "Test",
	Fields: graphql.InputObjectConfigFieldMap{
		"one": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"str": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"list": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(TestType),
		},
	},
})
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

		g.generateInput("Test", false, nil, ts)

		ex := []byte(`NewInputObject(graphql.InputObjectConfig{
	Name: "Test",
	Fields: graphql.InputObjectConfigFieldMap{
		"one": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
			DefaultValue: 1,
		},
		"str": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
			DefaultValue: "hello",
		},
		"list": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.Int),
			DefaultValue: []interface{}{1, 2, 3},
		},
	},
})
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

		g.generateDirective("Test", false, nil, ts)

		ex := []byte(`NewDirective(graphql.DirectiveConfig{
	Name: "Test",
	Locations: []string{
		"QUERY",
		"FIELD",
		"SCHEMA",
	},
})
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

		g.generateDirective("Test", false, nil, ts)

		ex := []byte(`NewDirective(graphql.DirectiveConfig{
	Name: "Test",
	Locations: []string{
		"QUERY",
		"FIELD",
		"SCHEMA",
	},
	Args: graphql.FieldConfigArgument{
		"one": &graphql.ArgumentConfig{
			Type: graphql.Int,
			DefaultValue: 1,
		},
		"str": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
			DefaultValue: "hello",
		},
		"list": &graphql.ArgumentConfig{
			Type: graphql.NewList(graphql.Int),
			DefaultValue: []interface{}{1, 2, 3},
		},
	},
})
`)

		gen.CompareBytes(subT, ex, g.Bytes())
	})
}

func TestGenerator_Generate(t *testing.T) {
	g := &Generator{}

	var b bytes.Buffer
	ctx := gen.WithContext(context.Background(), gen.TestCtx{Writer: &b})
	err := g.Generate(ctx, testDoc, `{"descriptions": true}`)
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

		err := g.Generate(ctx, testDoc, "")
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
	err = g.Generate(ctx, doc, `{"descriptions": true}`)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(b.String())

	// Output:
	// package main
	//
	// import "github.com/graphql-go/graphql"
	//
	// var QueryType = graphql.NewObject(graphql.ObjectConfig{
	// 	Name: "Query",
	//	Fields: graphql.Fields{
	//		"hello": &graphql.Field{
	//			Type: graphql.String,
	//			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return nil, nil },
	//		},
	//	},
	//	Description: "Query represents the queries this example provides.",
	// })
	//
}
