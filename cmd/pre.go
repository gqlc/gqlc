package cmd

import (
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/gqlc/cmd/plugin"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
func parseFlags(prefix *string, geners, opts map[string]compiler.Generator) func(*cobra.Command, []string) error {
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

			f := &oFlag{opts: make(map[string]interface{}), outDir: new(string)}
			pg := &plugin.Generator{Name: strings.TrimSuffix(name, "_out"), Prefix: *prefix}

			outFlag := *f
			outFlag.isOut = true
			cmd.Flags().Var(outFlag, name, "")
			geners[name] = pg

			optName := strings.Replace(name, "_out", "_opt", 1)
			optFlag := *f
			cmd.Flags().Var(optFlag, optName, "")
			opts[optName] = pg
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

// accumulateGens accumulates the selected code generators
func accumulateGens(prefix *string, geners, opts map[string]compiler.Generator, genOpts map[compiler.Generator]*oFlag) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed {
				return
			}

			var fg compiler.Generator
			g, isOpt := opts[f.Name]
			gen, exists := geners[f.Name]
			switch {
			case isOpt:
				fg = g
			case exists:
				fg = gen
			default:
				return
			}

			if genOpts[fg] != nil {
				return
			}

			of := f.Value.(oFlag)
			genOpts[fg] = &of
		})
		return nil
	}
}

var pluginTypes []string

func init() {
	rootCmd.Flags().StringSliceVarP(&pluginTypes, "types", "t", nil, "Provide .gql files containing types you wish to register with the compiler.")
}

// validatePluginTypes parses and validates any types given by the --types flag.
func validatePluginTypes(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	docs := make([]*ast.Document, 0, len(pluginTypes))
	for _, t := range pluginTypes {
		path := t
		if !filepath.IsAbs(t) {
			path = filepath.Join(wd, path)
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		doc, err := parser.ParseDoc(token.NewDocSet(), filepath.Base(path), f, 0)
		if err != nil {
			return err
		}
		docs = append(docs, doc)
	}

	errs := compiler.CheckTypes(docs, compiler.TypeCheckerFn(compiler.Validate))
	if len(errs) > 0 {
		// TODO: Compound errs
		return nil
	}

	return nil
}

// initGenDirs initializes each directory each generator will be outputting to.
func initGenDirs(fs afero.Fs, genOpts map[compiler.Generator]*oFlag) func(*cobra.Command, []string) error {
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
