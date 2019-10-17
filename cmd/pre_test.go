package cmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	preRunE := func(c *cli) func(*cobra.Command, []string) error {
		return parseFlags(c.pluginPrefix, &c.geners, c.fp)
	}

	testCli := newTestCli(nil, preRunE, noopRun)
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
			err := testCli.PreRunE(testCli.Command, testCase.Args)
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