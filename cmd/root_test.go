package cmd

//go:generate mockgen -package=cmd -destination=./mock_test.go github.com/gqlc/compiler CodeGenerator

import (
	"github.com/golang/mock/gomock"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testRunFn func(map[string]compiler.Generator, map[string]compiler.Generator, map[compiler.Generator]*oFlag) func(*cobra.Command, []string) error

func newTestCli(preRunE, runE testRunFn) *ccli {
	c := &ccli{
		root: &cobra.Command{
			Use:                "gqlc",
			DisableFlagParsing: true,
			Args:               cobra.MinimumNArgs(1),
		},
		geners:  make(map[string]compiler.Generator),
		opts:    make(map[string]compiler.Generator),
		genOpts: make(map[compiler.Generator]*oFlag),
	}
	c.root.PreRunE = preRunE(c.geners, c.opts, c.genOpts)
	c.root.RunE = runE(c.geners, c.opts, c.genOpts)
	c.root.Flags().StringSliceP("import_path", "I", []string{"."}, ``)

	return c
}

func noopRun(map[string]compiler.Generator, map[string]compiler.Generator, map[compiler.Generator]*oFlag) func(*cobra.Command, []string) error {
	return nil
}

var (
	testFs afero.Fs
	oneGql = `@import(paths: ["two.gql", "/usr/imports/six.gql", "four.gql"])

type Service implements Doc {
	t: Time
	obj: Obj
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

func TestPreRun(t *testing.T) {
	testCli := newTestCli(preRunRoot, noopRun)
	testGen := newMockGenerator(t)
	testCli.RegisterGenerator(testGen, "a_out", "A test generator.")
	testCli.RegisterGenerator(testGen, "b_out", "b_opt", "A second test generator")

	testCases := []struct {
		Name string
		Args []string
		Err  string
	}{
		{
			Name: "perfect",
			Args: []string{"--a_out", "aDir", "--b_out", "bDir", "test.gql"},
		},
		{
			Name: "missingFile(s)",
			Args: []string{},
			Err:  "requires at least 1 arg(s), only received 0",
		},
		{
			Name: "outWithOpts",
			Args: []string{"--b_out=a,b=b,c=1.5,d=false:bDir", "--b_opt=e,f=f,g=2", "test.gql"},
		},
		{
			Name: "justPlugin",
			Args: []string{"--plugin_out", ".", "test.gql"},
		},
		{
			Name: "pluginWithOpts",
			Args: []string{"--new_plugin_out=a,b=b,c=1.4,d=false:.", "--new_plugin_opt=e,f=f,g=2,h=false", "test.gql"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			err := testCli.root.PreRunE(testCli.root, testCase.Args)
			if err != nil {
				if err.Error() != testCase.Err {
					subT.Errorf("expected error: %s but got: %s", testCase.Err, err)
					return
				}
				return
			}
			if testCase.Err != "" {
				subT.Fail()
			}
		})
	}
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
			docs, err := parseInputFiles(testFs, testCase.ImportPaths, testCase.Args)
			if err != nil {
				subT.Error(err)
				return
			}

			if len(docs) != testCase.Len {
				subT.Fail()
				return
			}

			docMap := make(map[string]*ast.Document)
			for _, doc := range docs {
				docMap[doc.Name] = doc
			}

			for _, doc := range docMap {
				for _, direc := range doc.Directives {
					for _, arg := range direc.Args.Args {
						compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
						listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List
						paths := listLit.(*ast.ListLit_BasicList).BasicList
						for _, p := range paths.Values {
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

func TestRun(t *testing.T) {
	testCases := []struct {
		Name   string
		Args   []string
		expect func(g *MockCodeGenerator)
	}{
		{
			Name: "SingleWoImports",
			Args: []string{"gqlc", "/home/graphql/imports/thr.gql"},
			expect: func(g *MockCodeGenerator) {
				g.EXPECT().GenerateAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "SingleWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home/graphql/imports", "five.gql"},
			expect: func(g *MockCodeGenerator) {
				g.EXPECT().GenerateAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWoImports",
			Args: []string{"gqlc", "-I", "/home", "-I", "/home/graphql/imports", "thr.gql", "four.gql"},
			expect: func(g *MockCodeGenerator) {
				g.EXPECT().GenerateAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home", "-I=/home/graphql", "-I", "/home/graphql/imports", "one.gql", "five.gql"},
			expect: func(g *MockCodeGenerator) {
				g.EXPECT().GenerateAll(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			ctrl := gomock.NewController(subT)
			testGen := NewMockCodeGenerator(ctrl)

			testCli := newTestCli(preRunRoot, func(_, _ map[string]compiler.Generator, flags map[compiler.Generator]*oFlag) func(*cobra.Command, []string) error {
				return runRoot(testFs, flags)
			})
			testCli.RegisterGenerator(testGen, "test_out", "test_opt", "Test generator")

			err := testCli.Run(testCase.Args)
			if err != nil {
				subT.Error(err)
				return
			}
		})
	}
}
