package cmd

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"testing"
)

func newMockGenerator(t gomock.TestReporter) *MockGenerator {
	return NewMockGenerator(gomock.NewController(t))
}

func parseArgs(cmd *cobra.Command, args []string) error { return cmd.Flags().Parse(args) }

func TestCcli_RegisterGenerator(t *testing.T) {
	t.Run("singleOut", func(subT *testing.T) {
		name := "singleOut"

		testCli := newTestCli(noopPreRunE, noopRun)
		testCli.RegisterGenerator(newMockGenerator(subT), fmt.Sprintf("%s_out", name), "Test generator")

		err := parseArgs(testCli.Command, []string{fmt.Sprintf("--%s_out", name), "."})
		if err != nil {
			subT.Error(err)
			return
		}
		f := testCli.Flags().Lookup(fmt.Sprintf("%s_out", name))
		of := f.Value.(oFlag)
		if *of.outDir == "." {
			subT.Fail()
			return
		}
	})

	t.Run("singleOutWithOpts", func(subT *testing.T) {
		name := "singleOneWithOpts"

		testCli := newTestCli(noopPreRunE, noopRun)
		testCli.RegisterGenerator(newMockGenerator(subT), fmt.Sprintf("%s_out", name), "Test Generator")

		err := parseArgs(testCli.Command, []string{fmt.Sprintf("--%s_out=a,b=b,c=1.4,d=false:.", name)})
		if err != nil {
			subT.Error(err)
			return
		}

		f := testCli.Flags().Lookup(fmt.Sprintf("%s_out", name))
		of := f.Value.(oFlag)
		if *of.outDir != "." {
			subT.Fail()
			return
		}

		if !of.opts["a"].(bool) {
			subT.Fail()
			return
		}

		if of.opts["b"].(string) != "b" {
			subT.Fail()
			return
		}

		if of.opts["c"].(float64) != 1.4 {
			subT.Fail()
			return
		}

		if of.opts["d"].(bool) {
			subT.Fail()
			return
		}
	})

	t.Run("outAndOpt", func(subT *testing.T) {
		subT.Run("justOptsOnOut", func(triT *testing.T) {
			name := "justOptsOnOut"

			testCli := newTestCli(noopPreRunE, noopRun)
			testCli.RegisterGenerator(newMockGenerator(triT), fmt.Sprintf("%s_out", name), fmt.Sprintf("%s_opt", name), "Test Generator")

			err := parseArgs(testCli.Command, []string{fmt.Sprintf("--%s_out=a,b=b,c=1.4,d=false:.", name)})
			if err != nil {
				subT.Error(err)
				return
			}

			f := testCli.Flags().Lookup(fmt.Sprintf("%s_out", name))
			of := f.Value.(oFlag)
			if *of.outDir != "." {
				subT.Fail()
				return
			}

			if *of.outDir != "." {
				subT.Fail()
				return
			}

			if !of.opts["a"].(bool) {
				subT.Fail()
				return
			}

			if of.opts["b"].(string) != "b" {
				subT.Fail()
				return
			}

			if of.opts["c"].(float64) != 1.4 {
				subT.Fail()
				return
			}

			if of.opts["d"].(bool) {
				subT.Fail()
				return
			}
		})

		subT.Run("optsOnOutAndOpts", func(triT *testing.T) {
			name := "optsOnOutAndOpts"

			testCli := newTestCli(noopPreRunE, noopRun)
			testCli.RegisterGenerator(newMockGenerator(triT), fmt.Sprintf("%s_out", name), fmt.Sprintf("%s_opt", name), "Test Generator")

			err := parseArgs(testCli.Command, []string{fmt.Sprintf("--%s_out=a,b=b,c=1.4,d=false:.", name), fmt.Sprintf("--%s_opt=e,f=f,g=2,h=false", name)})
			if err != nil {
				subT.Error(err)
				return
			}

			f := testCli.Flags().Lookup(fmt.Sprintf("%s_out", name))
			of := f.Value.(oFlag)
			if *of.outDir != "." {
				subT.Fail()
				return
			}

			if *of.outDir != "." {
				subT.Fail()
				return
			}

			if !of.opts["a"].(bool) {
				subT.Fail()
				return
			}

			if of.opts["b"].(string) != "b" {
				subT.Fail()
				return
			}

			if of.opts["c"].(float64) != 1.4 {
				subT.Fail()
				return
			}

			if of.opts["d"].(bool) {
				subT.Fail()
				return
			}

			if !of.opts["e"].(bool) {
				subT.Fail()
				return
			}

			if of.opts["f"].(string) != "f" {
				subT.Fail()
				return
			}

			if of.opts["g"].(int64) != 2 {
				subT.Fail()
				return
			}

			if of.opts["h"].(bool) {
				subT.Fail()
				return
			}
		})
	})
}

func TestOutFlag_Set(t *testing.T) {

	t.Run("justDir", func(subT *testing.T) {
		f := &oFlag{opts: make(map[string]interface{}), outDir: new(string), isOut: true}

		arg := "testDir"
		err := f.Set(arg)
		if err != nil {
			subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
			return
		}

		if *f.outDir != arg {
			subT.Fail()
			return
		}

		subT.Run("absPath", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{}), outDir: new(string), isOut: true}

			arg := "/testDir"
			err := f.Set(arg)
			if err != nil {
				triT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if *f.outDir != arg {
				triT.Fail()
			}
		})

		subT.Run("relPath", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{}), outDir: new(string), isOut: true}

			arg := "testDir/a"
			err := f.Set(arg)
			if err != nil {
				triT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if *f.outDir != arg {
				triT.Fail()
			}
		})
	})

	t.Run("optsButNoDir", func(subT *testing.T) {
		f := &oFlag{opts: make(map[string]interface{})}

		arg := "testOpts:"
		err := f.Set(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
	})

	t.Run("malformedOpts", func(subT *testing.T) {
		f := &oFlag{opts: make(map[string]interface{})}

		arg := "testOpts=:"
		err := f.Set(arg)
		if err == nil {
			subT.Errorf("expected error from flag parsing: %s", arg)
			return
		}

		if err.Error() != "gqlc: unexpected character in generator option, testOpts, value: :" {
			subT.Error(err)
		}
	})

	t.Run("boolOpt", func(subT *testing.T) {
		f := &oFlag{opts: make(map[string]interface{})}

		arg := "testBoolOpt:"
		err := f.Set(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
		if !f.opts["testBoolOpt"].(bool) {
			subT.Fail()
		}

		arg = "testBoolOpt=false:"
		err = f.Set(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
		if f.opts["testBoolOpt"].(bool) {
			subT.Fail()
		}

		arg = "testBoolOpt=true:"
		err = f.Set(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
		if !f.opts["testBoolOpt"].(bool) {
			subT.Fail()
		}
	})

	t.Run("multiOpts", func(subT *testing.T) {

		subT.Run("multiInt", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{})}

			arg := "testInts=1,testInts=2,testInts=3:"
			err := f.Set(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testInts"].([]int64)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiFloat", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{})}

			arg := "testFloats=1.0,testFloats=2.0,testFloats=3.0:"
			err := f.Set(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testFloats"].([]float64)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiString", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{})}

			arg := `testStrings="1",testStrings="2",testStrings="3":`
			err := f.Set(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testStrings"].([]string)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiIdent", func(triT *testing.T) {
			f := &oFlag{opts: make(map[string]interface{})}

			arg := "testIdents=one,testIdents=two,testIdents=three:"
			err := f.Set(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testIdents"].([]string)) != 3 {
				triT.Fail()
			}
		})
	})

}
