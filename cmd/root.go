package cmd

import (
	"context"
	"fmt"
	"github.com/Zaba505/gqlc/compiler"
	"github.com/Zaba505/gqlc/graphql/ast"
	"github.com/Zaba505/gqlc/graphql/parser"
	"github.com/Zaba505/gqlc/graphql/token"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
	"strings"
)

// helper template funcs for rootCmd usage template
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

var rootCmd = &cobra.Command{
	Use:   "gqlc",
	Short: "A GraphQL IDL compiler",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1), // Make sure at least one file is provided.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" {
			return nil
		}

		// Validate file names
		for _, fileName := range args {
			ext := strings.TrimPrefix(filepath.Ext(fileName), ".")
			if ext != "gql" && ext != "graphql" {
				return fmt.Errorf("invalid file extension: %s", fileName)
			}
		}
		return nil
	},
	RunE:             runRoot,
	TraverseChildren: true,
}

func init() {
	cobra.AddTemplateFuncs(tmplFs)

	rootCmd.PersistentFlags().StringSliceP("import_path", "I", []string{"."}, `Specify the directory in which to search for
imports.  May be specified multiple times;
directories will be searched in order.  If not
given, the current working directory is used.`)
	rootCmd.Flags().BoolP("verbose", "v", false, "Output logging")
	rootCmd.SetUsageTemplate(`Usage:
	gqlc [command|flags] files{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{$flags := in .LocalFlags "_out"}}{{if gt (len $flags.FlagUsages) 0}}

Generator Flags:
{{$flags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{$flags = ex .LocalFlags "_out"}}{{if gt (len $flags.FlagUsages) 0}}

General Flags:
{{$flags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`)
}

func runRoot(cmd *cobra.Command, args []string) (err error) {
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
	schemas := make([]*ast.Document, 0, len(args))
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
