package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/gqlc/cmd/plugin"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func chainPreRunEs(preRunEs ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		for i := 0; i < len(preRunEs) && err == nil; i++ {
			err = preRunEs[i](cmd, args)
		}
		return
	}
}

// parseFlags parses the flags given and handles plugin flags
func parseFlags(prefix *string, geners *[]*genFlag, fp *fparser) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}

		// Parse flags and handle plugin flags
		var name string
		for _, a := range args {
			// Filter for output flags
			switch strings.Contains(a, "_out") {
			case false:
				continue
			case strings.Contains(a, ":"):
				ss := strings.Split(a, ":")
				name = ss[0][:strings.IndexRune(ss[0], '=')]
			default:
				name = a
			}

			// Trim "--" prefix
			name = name[2:]
			if f := cmd.Flags().Lookup(name); f != nil {
				continue
			}

			opts := make(map[string]interface{})
			cmd.Flags().Var(&genFlag{
				Generator: &plugin.Generator{Name: strings.TrimSuffix(name, "_out"), Prefix: *prefix},
				outDir:    new(string),
				opts:      opts,
				geners:    geners,
				fp:        fp,
			}, name, "")

			optName := strings.Replace(name, "_out", "_opt", 1)
			cmd.Flags().Var(&genOptFlag{opts: opts, fp: fp}, optName, "")
		}

		return cmd.Flags().Parse(args)
	}
}

// validateArgs validates the args given to the command
func validateArgs(cmd *cobra.Command, args []string) error {
	args = cmd.Flags().Args()

	// Validate args
	if err := cmd.ValidateArgs(args); err != nil {
		return err
	}

	// Validate file names
	for _, fileName := range args {
		ext := filepath.Ext(fileName)
		if ext != ".gql" && ext != ".graphql" {
			return fmt.Errorf("gqlc: invalid file extension: %s", fileName)
		}
	}

	return nil
}

func init() {
	rootCmd.Flags().StringSliceP("types", "t", nil, "Provide .gql files containing types you wish to register with the compiler.")
}

// validatePluginTypes parses and validates any types given by the --types flag.
func validatePluginTypes(fs afero.Fs) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		pluginTypes, _ := cmd.LocalFlags().GetStringSlice("types")
		if len(pluginTypes) == 0 {
			return nil
		}

		importPaths, err := cmd.Flags().GetStringSlice("import_path")
		if err != nil {
			return err
		}

		docs, err := parseInputFiles(fs, importPaths, pluginTypes)
		if err != nil {
			return err
		}

		docs, err = compiler.ReduceImports(docs)
		if err != nil {
			return err
		}

		errs := compiler.CheckTypes(docs, compiler.TypeCheckerFn(compiler.Validate))
		if len(errs) > 0 {
			// TODO: Compound errs
			return nil
		}

		return nil
	}
}

// initGenDirs initializes each directory each generator will be outputting to.
func initGenDirs(fs afero.Fs, genOpts []*genFlag) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		for _, genOpt := range genOpts {
			err = fs.MkdirAll(*genOpt.outDir, os.ModeDir)
			if err != nil {
				break
			}
		}
		return
	}
}
