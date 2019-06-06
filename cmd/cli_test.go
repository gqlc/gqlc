package cmd

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"testing"
	"text/scanner"
)

func newMockGenerator(t gomock.TestReporter) *MockGenerator {
	return NewMockGenerator(gomock.NewController(t))
}

type testRunFn func(afero.Fs, *[]*genFlag) func(*cobra.Command, []string) error

func newTestCli(fs afero.Fs, preRunE func(*cli) func(*cobra.Command, []string) error, runE testRunFn) *cli {
	c := &cli{
		Command: &cobra.Command{
			Use:                "gqlc",
			DisableFlagParsing: true,
			Args:               cobra.MinimumNArgs(1),
		},
		pluginPrefix: new(string),
		fp: &fparser{
			Scanner: new(scanner.Scanner),
		},
	}
	c.PreRunE = preRunE(c)
	c.RunE = runE(fs, &c.geners)
	c.Flags().StringSliceP("import_path", "I", []string{"."}, ``)

	return c
}

func noopPreRunE(*cli) func(*cobra.Command, []string) error {
	return func(*cobra.Command, []string) error { return nil }
}

func noopRun(afero.Fs, *[]*genFlag) func(*cobra.Command, []string) error { return nil }

func parseArgs(cmd *cobra.Command, args []string) error { return cmd.Flags().Parse(args) }

func TestCli_RegisterGenerator(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	testCases := []struct {
		Name   string
		Args   []string
		OutDir string
		Opts   map[string]interface{}
		Err    string
	}{
		{
			Name:   "NoOptsWithOut",
			Args:   []string{"--NoOptsWithOut_out", "."},
			OutDir: wd,
		},
		{
			Name:   "OptsOnOut",
			Args:   []string{"--OptsOnOut_out=a,b=b,c=1.4,d=false:."},
			OutDir: wd,
			Opts:   map[string]interface{}{"a": true, "b": "b", "c": 1.4, "d": false},
		},
		{
			Name:   "OptFlagAndOutFlagOpts",
			Args:   []string{"--OptFlagAndOutFlagOpts_out=a,b=b,c=1.4,d=false:.", "--OptFlagAndOutFlagOpts_opt=e,f=f,g=2,h=false,i"},
			OutDir: wd,
			Opts:   map[string]interface{}{"a": true, "b": "b", "c": 1.4, "d": false, "e": true, "f": "f", "g": int64(2), "h": false, "i": true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			name := fmt.Sprintf("%s_out", testCase.Name)

			testCli := newTestCli(nil, noopPreRunE, noopRun)
			testCli.RegisterGenerator(newMockGenerator(subT), name, fmt.Sprintf("%s_opt", testCase.Name), "Test Generator")

			err := parseArgs(testCli.Command, testCase.Args)
			if err != nil && testCase.Err == "" {
				subT.Errorf("unexpected error from arg parsing: %s:%s", testCase.Args, err)
				return
			}
			if testCase.Err != "" {
				if err == nil {
					subT.Errorf("expected error: %s", testCase.Err)
					return
				}

				if err.Error() != testCase.Err {
					subT.Logf("mismatched errors: %s:%s", err, testCase.Err)
					subT.Fail()
				}
				return
			}

			f := testCli.Flags().Lookup(name).Value.(*genFlag)
			if testCase.OutDir != *f.outDir {
				subT.Logf("mismatched output dirs: %s:%s", *f.outDir, testCase.OutDir)
				subT.Fail()
				return
			}

			if len(f.opts) != len(testCase.Opts) {
				subT.Fail()
			}

			compare(subT, f.opts, testCase.Opts)
		})
	}
}

func TestCli_Run(t *testing.T) {
	preRunE := func(c *cli) func(*cobra.Command, []string) error {
		return chainPreRunEs(
			parseFlags(c.pluginPrefix, &c.geners, c.fp),
			validateArgs,
			validatePluginTypes(testFs),
			initGenDirs(testFs, c.geners),
		)
	}

	c := newTestCli(testFs, preRunE, root)

	testCases := []struct {
		Name   string
		Args   []string
		expect func(g *MockGenerator)
	}{
		{
			Name: "SingleWoImports",
			Args: []string{"gqlc", "/home/graphql/imports/thr.gql"},
			expect: func(g *MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "SingleWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home/graphql/imports", "five.gql"},
			expect: func(g *MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWoImports",
			Args: []string{"gqlc", "-I", "/home", "-I", "/home/graphql/imports", "thr.gql", "four.gql"},
			expect: func(g *MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			Name: "MultiWImports",
			Args: []string{"gqlc", "-I", "/usr/imports", "-I", "/home", "-I=/home/graphql", "-I", "/home/graphql/imports", "one.gql", "five.gql"},
			expect: func(g *MockGenerator) {
				g.EXPECT().Generate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			if err := c.Run(testCase.Args); err != nil {
				subT.Error(err)
				return
			}
		})
	}
}

func TestCli_Run_Recover(t *testing.T) {
	f := func() { panic("test") }

	c := newTestCli(nil, noopPreRunE, func(afero.Fs, *[]*genFlag) func(*cobra.Command, []string) error {
		return func(*cobra.Command, []string) error {
			f() // the panic call can't go here cuz go vet can detect that the return won't be reached
			return nil
		}
	})

	err := c.Run([]string{"test", ""})
	perr, ok := err.(*panicErr)
	if !ok {
		t.Error(err)
		return
	}
	if perr.Err.Error() != `"test"` {
		t.Fail()
	}
}

func compare(t *testing.T, out, ex map[string]interface{}) {
	match := true
	var missing []string
	for k, outVal := range out {
		exVal, exists := ex[k]
		if !exists {
			missing = append(missing, k)
		}

		switch v := outVal.(type) {
		case int64, float64, string, bool:
			match = v == exVal
		case []int64:
			_, match = exVal.([]int64)
			if !match {
				break
			}

			exSlice := exVal.([]int64)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []float64:
			_, match = exVal.([]float64)
			if !match {
				break
			}

			exSlice := exVal.([]float64)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []string:
			_, match = exVal.([]string)
			if !match {
				break
			}

			exSlice := exVal.([]string)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		case []bool:
			_, match = exVal.([]bool)
			if !match {
				break
			}

			exSlice := exVal.([]bool)
			match = len(exSlice) == len(v)
			if !match {
				break
			}

			for i := range exSlice {
				if match = v[i] == exSlice[i]; !match {
					break
				}
			}
		default:
			match = false
		}

		if !match {
			t.Fail()
			t.Logf("mismatched values for key, %s: %v:%v", k, outVal, exVal)
		}

		delete(ex, k)
	}

	for _, k := range missing {
		t.Logf("key found in output and not in expected: %s", k)
	}

	for k := range ex {
		t.Logf("expected key: %s", k)
	}
}
