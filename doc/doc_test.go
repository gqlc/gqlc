package doc

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func compareBytes(t *testing.T, ex, out []byte) {
	if bytes.EqualFold(out, ex) {
		return
	}

	line := 1
	for i, b := range out {
		if b == '\n' {
			line++
		}

		if ex[i] != b {
			t.Fatalf("expected: %s, but got: %s, %d:%d", string(ex[i]), string(b), i, line)
		}
	}
}

var (
	// Flags are used here to allow for the input/output files to be changed during dev
	// One use case of changing the files is to examine how Generate scales through the benchmark
	//
	gqlFileName = flag.String("gqlFile", "test.gql", "Specify a .gql file to use a input for testing.")
	exDocName   = flag.String("expectedFile", "test.md", "Specify a file which is the expected generator output from the given .gql file.")

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

type testCtx struct {
	io.Writer
}

func (ctx testCtx) Open(filename string) (io.WriteCloser, error) { return ctx, nil }

func (ctx testCtx) Close() error { return nil }

func TestAddContent(t *testing.T) {
	mask := schemaType | scalarType | objectType | interType | unionType | enumType | inputType | directiveType | extendType

	testCases := []struct {
		Name string
		C    []struct {
			name  string
			count int
			typ   declType
		}
		Total int
	}{
		{
			Name: "SingleType",
			C: []struct {
				name  string
				count int
				typ   declType
			}{
				{
					name: "Test",
					typ:  scalarType,
				},
			},
			Total: 2,
		},
		{
			Name: "MultiSameType",
			C: []struct {
				name  string
				count int
				typ   declType
			}{
				{
					name: "A",
					typ:  scalarType,
				},
				{
					name: "B",
					typ:  scalarType,
				},
				{
					name: "C",
					typ:  scalarType,
				},
			},
			Total: 4,
		},
		{
			Name: "ManyTypes",
			C: []struct {
				name  string
				count int
				typ   declType
			}{
				{
					name: "A",
					typ:  scalarType,
				},
				{
					name: "B",
					typ:  scalarType,
				},
				{
					name: "C",
					typ:  scalarType,
				},
				{
					name: "A",
					typ:  objectType,
				},
				{
					name: "B",
					typ:  objectType,
				},
				{
					name: "C",
					typ:  objectType,
				},
			},
			Total: 8,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			toc := make([]struct {
				name  string
				count int
			}, 0, len(testCase.C))
			opts := &Options{
				toc: &toc,
			}

			tmask := mask
			for _, c := range testCase.C {
				tmask = opts.addContent(c.name, c.count, c.typ, tmask)
			}

			if len(*opts.toc) != testCase.Total {
				fmt.Println(*opts.toc)
				subT.Fail()
			}
		})
	}

}

func TestToC(t *testing.T) {
	testCases := []struct {
		Name string
		ToC  []struct {
			name  string
			count int
		}
		Ex []byte
	}{
		{
			Name: "SingleSection",
			ToC: []struct {
				name  string
				count int
			}{
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
			},
			Ex: []byte(`# Test
*This was generated by gqlc.*

## Table of Contents
- [Scalars](#Scalars)
	* [Int](#Int)
	* [Float](#Float)
	* [String](#String)

`),
		},
		{
			Name: "MultipleSections",
			ToC: []struct {
				name  string
				count int
			}{
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
				{
					name: object,
				},
				{
					name: "Person",
				},
				{
					name: "Hero",
				},
				{
					name: "Jedi",
				},
				{
					name: inter,
				},
				{
					name: "Node",
				},
				{
					name: "Connection",
				},
			},
			Ex: []byte(`# Test
*This was generated by gqlc.*

## Table of Contents
- [Scalars](#Scalars)
	* [Int](#Int)
	* [Float](#Float)
	* [String](#String)
- [Objects](#Objects)
	* [Person](#Person)
	* [Hero](#Hero)
	* [Jedi](#Jedi)
- [Interfaces](#Interfaces)
	* [Node](#Node)
	* [Connection](#Connection)

`),
		},
		{
			Name: "SingleSectionWithExts",
			ToC: []struct {
				name  string
				count int
			}{
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
				{
					name: extend,
				},
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
			},
			Ex: []byte(`# Test
*This was generated by gqlc.*

## Table of Contents
- [Scalars](#Scalars)
	* [Int](#Int)
	* [Float](#Float)
	* [String](#String)
- [Extensions](#Extensions)
	* [Scalar Extensions](#Scalar-Extensions)
		- [Int Extension](#Int-Extension)
		- [Float Extension](#Float-Extension)
		- [String Extension](#String-Extension)

`),
		},
		{
			Name: "MultiSectionsWithExts",
			ToC: []struct {
				name  string
				count int
			}{
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
				{
					name: object,
				},
				{
					name: "Person",
				},
				{
					name: "Hero",
				},
				{
					name: "Jedi",
				},
				{
					name: inter,
				},
				{
					name: "Node",
				},
				{
					name: "Connection",
				},
				{
					name: extend,
				},
				{
					name: scalar,
				},
				{
					name: "Int",
				},
				{
					name: "Float",
				},
				{
					name: "String",
				},
				{
					name: object,
				},
				{
					name: "Person",
				},
				{
					name: "Hero",
				},
				{
					name: "Jedi",
				},
				{
					name: inter,
				},
				{
					name: "Node",
				},
				{
					name: "Connection",
				},
			},
			Ex: []byte(`# Test
*This was generated by gqlc.*

## Table of Contents
- [Scalars](#Scalars)
	* [Int](#Int)
	* [Float](#Float)
	* [String](#String)
- [Objects](#Objects)
	* [Person](#Person)
	* [Hero](#Hero)
	* [Jedi](#Jedi)
- [Interfaces](#Interfaces)
	* [Node](#Node)
	* [Connection](#Connection)
- [Extensions](#Extensions)
	* [Scalar Extensions](#Scalar-Extensions)
		- [Int Extension](#Int-Extension)
		- [Float Extension](#Float-Extension)
		- [String Extension](#String-Extension)
	* [Object Extensions](#Object-Extensions)
		- [Person Extension](#Person-Extension)
		- [Hero Extension](#Hero-Extension)
		- [Jedi Extension](#Jedi-Extension)
	* [Interface Extensions](#Interface-Extensions)
		- [Node Extension](#Node-Extension)
		- [Connection Extension](#Connection-Extension)

`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			opts := &Options{
				Title: "Test",
				toc:   &testCase.ToC,
			}

			var b bytes.Buffer
			writeToC(&b, opts)
			compareBytes(subT, testCase.Ex, b.Bytes())
		})
	}
}

func TestFields(t *testing.T) {
	g := new(Generator)

	testCases := []struct {
		Name   string
		Fields []*ast.Field
		Ex     []byte
	}{
		{
			Name: "JustFields",
			Fields: []*ast.Field{
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
			Ex: []byte(`- one **(Int)**
- str **(String)**
- list **([[Test](#Test)])**
`),
		},
		{
			Name: "WithDescriptions",
			Fields: []*ast.Field{
				{
					Name: &ast.Ident{Name: "one"},
					Doc: &ast.DocGroup{
						List: []*ast.DocGroup_Doc{
							{Text: "one is a Int."},
						},
					},
					Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Int"}},
				},
				{
					Name: &ast.Ident{Name: "str"},
					Doc: &ast.DocGroup{
						List: []*ast.DocGroup_Doc{
							{Text: "str is a String."},
						},
					},
					Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
				},
				{
					Name: &ast.Ident{Name: "list"},
					Doc: &ast.DocGroup{
						List: []*ast.DocGroup_Doc{
							{Text: "list is a List."},
						},
					},
					Type: &ast.Field_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
				},
			},
			Ex: []byte(`- one **(Int)**

	one is a Int.
- str **(String)**

	str is a String.
- list **([[Test](#Test)])**

	list is a List.
`),
		},
		{
			Name: "WithArgs",
			Fields: []*ast.Field{
				{
					Name: &ast.Ident{Name: "one"},
					Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "Int"}},
					Args: &ast.InputValueList{
						List: []*ast.InputValue{
							{
								Name: &ast.Ident{Name: "toNumber"},
								Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
							},
						},
					},
				},
				{
					Name: &ast.Ident{Name: "str"},
					Type: &ast.Field_Ident{Ident: &ast.Ident{Name: "String"}},
					Args: &ast.InputValueList{
						List: []*ast.InputValue{
							{
								Name: &ast.Ident{Name: "toString"},
								Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
							},
						},
					},
				},
				{
					Name: &ast.Ident{Name: "list"},
					Type: &ast.Field_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
					Args: &ast.InputValueList{
						List: []*ast.InputValue{
							{
								Name: &ast.Ident{Name: "first"},
								Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
							},
							{
								Name: &ast.Ident{Name: "after"},
								Type: &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
							},
						},
					},
				},
			},
			Ex: []byte(`- one **(Int)**

	*Args*:
	- toNumber **(String)**
- str **(String)**

	*Args*:
	- toString **(Int)**
- list **([[Test](#Test)])**

	*Args*:
	- first **(Int)**
	- after **(String)**
`),
		},
	}

	var testBuf bytes.Buffer
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			g.Lock()
			defer g.Unlock()
			g.Reset()

			g.generateFields(testCase.Fields, &testBuf)

			compareBytes(subT, testCase.Ex, g.Bytes())
		})
	}
}

func TestArgs(t *testing.T) {
	g := new(Generator)

	testCases := []struct {
		Name string
		Args []*ast.InputValue
		Ex   []byte
	}{
		{
			Name: "WithDefaults",
			Args: []*ast.InputValue{
				{
					Name:    &ast.Ident{Name: "one"},
					Type:    &ast.InputValue_Ident{Ident: &ast.Ident{Name: "Int"}},
					Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{Kind: token.Token_INT, Value: "1"}},
				},
				{
					Name:    &ast.Ident{Name: "str"},
					Type:    &ast.InputValue_Ident{Ident: &ast.Ident{Name: "String"}},
					Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{Kind: token.Token_STRING, Value: "\"hello\""}},
				},
				{
					Name: &ast.Ident{Name: "list"},
					Type: &ast.InputValue_List{List: &ast.List{Type: &ast.List_Ident{Ident: &ast.Ident{Name: "Test"}}}},
					Default: &ast.InputValue_CompositeLit{CompositeLit: &ast.CompositeLit{
						Value: &ast.CompositeLit_ListLit{ListLit: &ast.ListLit{
							List: &ast.ListLit_BasicList{
								BasicList: &ast.ListLit_Basic{
									Values: []*ast.BasicLit{
										{Value: "1"},
										{Value: "2"},
										{Value: "3"},
									},
								},
							},
						}},
					}},
				},
			},
			Ex: []byte("- one **(Int)**\n\n" +
				"	*Default Value*: `1`\n" +
				"- str **(String)**\n\n" +
				"	*Default Value*: `\"hello\"`\n" +
				"- list **([[Test](#Test)])**\n\n" +
				"	*Default Value*: `[1, 2, 3]`\n"),
		},
	}

	var testBuf bytes.Buffer
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			g.Lock()
			defer g.Unlock()
			g.Reset()

			g.generateArgs(testCase.Args, &testBuf)

			compareBytes(subT, testCase.Ex, g.Bytes())
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	var b bytes.Buffer
	gen := new(Generator)
	ctx := compiler.WithContext(context.Background(), testCtx{Writer: &b})
	err := gen.Generate(ctx, testDoc, "")
	if err != nil {
		t.Error(err)
		return
	}

	// Compare generated output to golden output
	ex, err := ioutil.ReadAll(exDoc)
	if err != nil {
		t.Error(err)
		return
	}

	compareBytes(t, ex, b.Bytes())
}

func BenchmarkGenerator_Generate(b *testing.B) {
	var buf bytes.Buffer
	g := new(Generator)
	ctx := compiler.WithContext(context.Background(), testCtx{Writer: &buf})

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
		return // Handle error
	}

	var b bytes.Buffer
	ctx := compiler.WithContext(context.Background(), &testCtx{Writer: &b}) // Pass in an actual
	err = g.Generate(ctx, doc, `{"title": "Example Documentation"}`)
	if err != nil {
		return // Handle err
	}

	fmt.Println(b.String())

	// Output:
	// # Example Documentation
	// *This was generated by gqlc.*
	//
	// ## Table of Contents
	// - [Schema](#Schema)
	// - [Objects](#Objects)
	// 	* [Query](#Query)
	//
	// ## Schema
	//
	// *Root Operations*:
	// - query **([Query](#Query))**
	//
	// ## Objects
	//
	// ### Query
	// Query represents the queries this example provides.
	//
	// *Fields*:
	// - hello **(String)**
}
