package cmd

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestFilterFlags(t *testing.T) {
	fs := new(pflag.FlagSet)
	fs.StringP("a_out", "a", "", "")
	fs.StringP("b_out", "b", "", "")
	fs.String("a_opt", "", "")
	fs.String("b_opt", "", "")

	inFlags := map[string]struct{}{"a_out": {}, "b_out": {}}
	infs := filterFlags(fs, "_out", true)
	infs.VisitAll(func(f *pflag.Flag) {
		delete(inFlags, f.Name)
	})
	if len(inFlags) > 0 {
		t.Fail()
		return
	}

	exFlags := map[string]struct{}{"a_opt": {}, "b_opt": {}}
	exfs := filterFlags(fs, "_out", false)
	exfs.VisitAll(func(f *pflag.Flag) {
		delete(exFlags, f.Name)
	})
	if len(exFlags) > 0 {
		t.Fail()
	}
}
