package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
)

var (
	testFs afero.Fs
	oneGql = `@import(paths: ["two.gql", "/usr/imports/six.gql", "four.gql"])

type Service implements Doc {
	t: Time
	obj: Obj
	v: Version
}`
	twoGql = `@import(paths: ["./thr.gql"])

interface Doc {
	v: Version
}`
	thrGql  = `scalar Version`
	fourGql = `scalar Time`
	fiveGql = `@import(paths: ["six.gql", "/home/graphql/imports/two.gql"])

type T implements Doc {
	v: Version
	obj: Obj
}`
	sixGql = `@import(paths: ["/home/graphql/imports/thr.gql"])

type Obj {
	v: Version
}`
)

func TestMain(m *testing.M) {
	// Set up test fs
	testFs = afero.NewMemMapFs()
	testFs.MkdirAll("/home/graphql/imports", 0755)
	testFs.MkdirAll("/usr/imports", 0755)

	afero.WriteFile(testFs, "/home/graphql/one.gql", []byte(oneGql), 0644)
	afero.WriteFile(testFs, "/home/graphql/imports/two.gql", []byte(twoGql), 0644)
	afero.WriteFile(testFs, "/home/graphql/imports/thr.gql", []byte(thrGql), 0644)
	afero.WriteFile(testFs, "/home/four.gql", []byte(fourGql), 0644)
	afero.WriteFile(testFs, "/usr/imports/five.gql", []byte(fiveGql), 0644)
	afero.WriteFile(testFs, "/usr/imports/six.gql", []byte(sixGql), 0644)

	os.Exit(m.Run())
}

func TestParseInputFiles(t *testing.T) {
	// Create test cases
	testCases := []struct {
		Name        string
		ImportPaths []string
		Args        []string
		Len         int
	}{
		{
			Name: "SingleWithAbs",
			Len:  1,
			Args: []string{"/home/graphql/imports/thr.gql"},
		},
		{
			Name:        "SingleWithRel",
			Len:         1,
			Args:        []string{"./thr.gql"},
			ImportPaths: []string{"/home/graphql/imports"},
		},
		{
			Name:        "SingleWithIpath",
			Len:         1,
			Args:        []string{"thr.gql"},
			ImportPaths: []string{"/home/graphql/imports"},
		},
		{
			Name:        "MSingle",
			Len:         2,
			Args:        []string{"thr.gql", "four.gql"},
			ImportPaths: []string{"/home/graphql/imports", "/home"},
		},
		{
			Name:        "TreeIPath",
			Len:         5,
			Args:        []string{"one.gql"},
			ImportPaths: []string{"/home", "/home/graphql", "/home/graphql/imports", "/usr/imports"},
		},
		{
			Name: "TreeAllArgs",
			Len:  5,
			Args: []string{"/home/graphql/one.gql", "/home/graphql/imports/two.gql", "/home/graphql/imports/thr.gql", "/home/four.gql", "/usr/imports/six.gql"},
		},
		{
			Name:        "MTreeIPaths",
			Len:         6,
			Args:        []string{"one.gql", "five.gql"},
			ImportPaths: []string{"/home", "/home/graphql", "/home/graphql/imports", "/usr/imports"},
		},
		{
			Name: "MTreeAllArgs",
			Len:  6,
			Args: []string{"/home/graphql/one.gql", "/home/graphql/imports/two.gql", "/home/graphql/imports/thr.gql", "/home/four.gql", "/usr/imports/five.gql", "/usr/imports/six.gql"},
		},
	}

	// Run test cases
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			docMap := make(map[string]*ast.Document, len(testCase.Args))
			err := parseInputFiles(testFs, token.NewDocSet(), docMap, testCase.ImportPaths, testCase.Args...)
			if err != nil {
				subT.Error(err)
				return
			}

			if len(docMap) != testCase.Len {
				subT.Fail()
				return
			}

			for _, doc := range docMap {
				for _, direc := range doc.Directives {
					for _, arg := range direc.Args.Args {
						compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
						listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List

						var paths []*ast.BasicLit
						switch v := listLit.(type) {
						case *ast.ListLit_BasicList:
							paths = append(paths, v.BasicList.Values...)
						case *ast.ListLit_CompositeList:
							cpaths := v.CompositeList.Values
							paths = make([]*ast.BasicLit, len(cpaths))
							for i, c := range cpaths {
								paths[i] = c.Value.(*ast.CompositeLit_BasicLit).BasicLit
							}
						}

						for _, p := range paths {
							iPath := strings.Trim(p.Value, "\"")
							iName := filepath.Base(iPath)
							if _, exists := docMap[iName]; !exists {
								subT.Fail()
								return
							}
						}
					}
				}
			}
		})
	}
}

func TestRoot(t *testing.T) {
	testCases := []struct {
		Name   string
		IPaths []string
		Args   []string
		expect func(g *gen.MockGenerator)
	}{
		{
			Name: "SingleWoImports",
			Args: []string{"/home/graphql/imports/thr.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name:   "SingleWImports",
			IPaths: []string{"/usr/imports", "/home/graphql/imports"},
			Args:   []string{"five.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name:   "MultiWoImports",
			IPaths: []string{"/home", "/home/graphql/imports"},
			Args:   []string{"thr.gql", "four.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
		},
		{
			Name:   "MultiWImports",
			IPaths: []string{"/usr/imports", "/home", "/home/graphql", "/home/graphql/imports"},
			Args:   []string{"one.gql", "five.gql"},
			expect: func(g *gen.MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			g := newMockGenerator(subT)
			testCase.expect(g)

			geners := []generator{{
				Generator: g,
			}}

			err := root(testFs, geners, testCase.IPaths, testCase.Args...)
			if err != nil {
				subT.Error(err)
				return
			}
		})
	}
}

func TestRoot_AutoImplInterfaces(t *testing.T) {
	autoImplInterfacesGql := `
interface Iterator {
	next: Int
}

type Bytes implements Iterator {
	asString: String!
}
	`
	afero.WriteFile(testFs, "/home/graphql/auto_impl_interfaces.gql", []byte(autoImplInterfacesGql), 0644)

	g := newMockGenerator(t)
	g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	geners := []generator{{
		Generator: g,
	}}

	err := root(testFs, geners, nil, "/home/graphql/auto_impl_interfaces.gql")
	if err != nil {
		t.Error(err)
		return
	}
}
