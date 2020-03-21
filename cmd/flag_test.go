package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"text/scanner"
)

func TestGenFlag_Set(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("unexpected error when getting wd: %s", err)
		return
	}

	testCases := []struct {
		Name   string
		Arg    string
		OutDir string
		Opts   map[string]interface{}
		Err    string
	}{
		{
			Name:   "AbsPathDir",
			Arg:    "/testdir",
			OutDir: "/testdir",
		},
		{
			Name:   "RelPathDir",
			Arg:    "testdir/a",
			OutDir: "testdir/a",
		},
		{
			Name:   "RelPathDir-2",
			Arg:    "../testdir/a",
			OutDir: filepath.Join(wd, "../testdir/a"),
		},
		{
			Name: "NoDir",
			Arg:  "testOpt:",
			Opts: map[string]interface{}{"testOpt": true},
		},
		{
			Name: "MalformedOpts",
			Arg:  "testOpts=:",
			Err:  "gqlc: unexpected character in generator option, testOpts, value: :",
		},
		{
			Name: "FalseBoolOpt",
			Arg:  "testBoolOpt=false:",
			Opts: map[string]interface{}{"testBoolOpt": false},
		},
		{
			Name: "TrueBoolOpt",
			Arg:  "testBoolOpt=true:",
			Opts: map[string]interface{}{"testBoolOpt": true},
		},
		{
			Name: "MultiInt",
			Arg:  "testInts=1,testInts=2,testInts=3:",
			Opts: map[string]interface{}{"testInts": []int64{1, 2, 3}},
		},
		{
			Name: "MultiFloat",
			Arg:  "testFloats=1.0,testFloats=2.0,testFloats=3.0:",
			Opts: map[string]interface{}{"testFloats": []float64{1, 2, 3}},
		},
		{
			Name: "MultiString",
			Arg:  `testStrings="1",testStrings="2",testStrings="3":`,
			Opts: map[string]interface{}{"testStrings": []string{`"1"`, `"2"`, `"3"`}},
		},
		{
			Name: "MultiBool",
			Arg:  "testBools=true,testBools=false,testBools=true",
			Opts: map[string]interface{}{"testBools": []bool{true, false, true}},
		},
		{
			Name: "MultiIdent",
			Arg:  "testIdents=one,testIdents=two,testIdents=three:",
			Opts: map[string]interface{}{"testIdents": []string{"one", "two", "three"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			f := &genFlag{
				opts:   make(map[string]interface{}),
				outDir: new(string),
				geners: new([]*genFlag),
				fp:     &fparser{Scanner: new(scanner.Scanner)},
			}

			err := f.Set(testCase.Arg)
			if err != nil && testCase.Err == "" {
				subT.Errorf("unexpected error from flag parsing: %s:%s", testCase.Arg, err)
				return
			}
			if testCase.Err != "" {
				if err == nil {
					subT.Errorf("expected error: %s", testCase.Err)
					return
				}

				if err.Error() != testCase.Err {
					subT.Fail()
				}
				return
			}

			if testCase.OutDir != *f.outDir {
				subT.Logf("mismatched outdirs: %s:%s", testCase.OutDir, *f.outDir)
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
