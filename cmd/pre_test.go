package cmd

import (
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func TestValidateArgs(t *testing.T) {
	cmd := &cobra.Command{}

	err := validateFilenames(cmd, []string{"test.txt"})
	if err == nil {
		t.Fail()
		return
	}
}

func TestValidatePluginTypes(t *testing.T) {
	fs := afero.NewMemMapFs()
	f, err := fs.OpenFile("test.gql", os.O_CREATE, 755)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = f.Write([]byte(thrGql))
	if err != nil {
		t.Error(err)
		return
	}

	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("types", []string{"test.gql"}, "")
	cmd.Flags().StringSlice("import_path", []string{"."}, "")

	err = validatePluginTypes(fs)(cmd, nil)
	if err != nil {
		t.Error(err)
		return
	}
}

var (
	aDir = "a"
	bDir = "b"
)

func TestInitGenDirs(t *testing.T) {
	fs := afero.NewMemMapFs()
	gens := []string{aDir, bDir}

	err := initGenDirs(fs, &gens)(nil, nil)
	if err != nil {
		t.Error(err)
		return
	}

	b, err := afero.DirExists(fs, "a")
	if !b || err != nil {
		t.Fail()
		return
	}

	b, err = afero.DirExists(fs, "b")
	if !b || err != nil {
		t.Fail()
		return
	}
}
