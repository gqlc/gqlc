package main

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
	"testing"
)

type testGenerator struct{}

func (g *testGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

func (g *testGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}

func TestOutFlag_Parse(t *testing.T) {

	t.Run("justDir", func(subT *testing.T) {
		f := &outFlag{opts: make(map[string]interface{})}

		arg := "testDir"
		err := f.Parse(arg)
		if err != nil {
			subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
			return
		}

		if f.outDir != "testDir" {
			subT.Fail()
		}
	})

	t.Run("optsButNoDir", func(subT *testing.T) {
		f := &outFlag{opts: make(map[string]interface{})}

		arg := "testOpts:"
		err := f.Parse(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
	})

	t.Run("malformedOpts", func(subT *testing.T) {
		f := &outFlag{opts: make(map[string]interface{})}

		arg := "testOpts=:"
		err := f.Parse(arg)
		if err == nil {
			subT.Errorf("expected error from flag parsing: %s", arg)
			return
		}

		if err.Error() != "gqlc: malformed generator option: testOpts" {
			subT.Error(err)
		}
	})

	t.Run("boolOpt", func(subT *testing.T) {
		f := &outFlag{opts: make(map[string]interface{})}

		arg := "testBoolOpt:"
		err := f.Parse(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
		if !f.opts["testBoolOpt"].(bool) {
			subT.Fail()
		}

		arg = "testBoolOpt=false:"
		err = f.Parse(arg)
		if err != nil {
			subT.Errorf("expected error from flag parsing: %s:%s", arg, err)
			return
		}
		if f.opts["testBoolOpt"].(bool) {
			subT.Fail()
		}

		arg = "testBoolOpt=true:"
		err = f.Parse(arg)
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
			f := &outFlag{opts: make(map[string]interface{})}

			arg := "testInts=1,testInts=2,testInts=3:"
			err := f.Parse(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testInts"].([]int64)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiFloat", func(triT *testing.T) {
			f := &outFlag{opts: make(map[string]interface{})}

			arg := "testFloats=1.0,testFloats=2.0,testFloats=3.0:"
			err := f.Parse(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testFloats"].([]float64)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiString", func(triT *testing.T) {
			f := &outFlag{opts: make(map[string]interface{})}

			arg := `testStrings="1",testStrings="2",testStrings="3":`
			err := f.Parse(arg)
			if err != nil {
				subT.Errorf("unexpected error from flag parsing: %s:%s", arg, err)
				return
			}

			if len(f.opts["testStrings"].([]string)) != 3 {
				triT.Fail()
			}
		})

		subT.Run("multiIdent", func(triT *testing.T) {
			f := &outFlag{opts: make(map[string]interface{})}

			arg := "testIdents=one,testIdents=two,testIdents=three:"
			err := f.Parse(arg)
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
