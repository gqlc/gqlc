package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gqlc/compiler"
	"gqlc/sl/file"
	"gqlc/sl/parser"
	"gqlc/sl/token"
	"os"
	"path/filepath"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:              "gqlc",
	Short:            "A GraphQL IDL compiler",
	Long:             ``,
	RunE:             runRoot,
	TraverseChildren: true,
}

var tmplFs = map[string]interface{}{
	"in": func(set *pflag.FlagSet, key string) *pflag.FlagSet {
		fs := new(pflag.FlagSet)
		set.VisitAll(func(flag *pflag.Flag) {
			if strings.Contains(flag.Name, key) {
				fs.AddFlag(flag)
			}
		})
		return fs
	},
	"ex": func(set *pflag.FlagSet, key string) *pflag.FlagSet {
		fs := new(pflag.FlagSet)
		set.VisitAll(func(flag *pflag.Flag) {
			if !strings.Contains(flag.Name, key) {
				fs.AddFlag(flag)
			}
		})
		return fs
	},
}

func init() {
	cobra.AddTemplateFuncs(tmplFs)

	rootCmd.Flags().StringP("schema_path", "I", ".", `Specify the directory in which to search for
imports.  May be specified multiple times;
directories will be searched in order.  If not
given, the current working directory is used.`)
	rootCmd.Flags().BoolP("verbose", "v", false, "Output for info")
	rootCmd.SetUsageTemplate(`Usage:
	gqlc [flags] files

Generator Flags:{{$flags := in .LocalFlags "_out"}}
{{$flags.FlagUsages | trimTrailingWhitespaces}}

General Flags:{{$flags = ex .LocalFlags "_out"}}
{{$flags.FlagUsages | trimTrailingWhitespaces}}
`)

	// TODO: Add sub commands to template
}

func runRoot(cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 {
		return errors.New("no files provided")
	}

	// Validate file names
	for _, fileName := range args {
		ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
		if ext != "gql" && ext != "graphql" {
			return fmt.Errorf("invalid file extension: %s", fileName)
		}
	}

	// Accumulate selected code generators
	var mode parser.Mode
	var gs []compiler.CodeGenerator
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}

		gen, exists := gens[f.Name]
		if exists {
			gs = append(gs, gen)
			if f.Name == "doc_out" {
				mode = parser.ParseComments
			}
		}
	})

	// Parse files
	schemas := make([]*file.Descriptor, 0, len(args))
	fset := token.NewFileSet()
	for _, filename := range args {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		schema, err := parser.ParseFile(fset, filename, f, mode)
		if err != nil {
			return err
		}

		schemas = append(schemas, schema)
	}

	// Run code generators
	for _, g := range gs {
		for _, schema := range schemas {
			err = g.Generate(context.TODO(), schema, "") // TODO: Replace context
			if err != nil {
				return
			}
		}
	}

	return
}
