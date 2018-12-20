package main

import (
	"context"
	"fmt"
	"github.com/Zaba505/gqlc/cmd/gqlc/util"
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
	Long: `gqlc is a multi-language GraphQL implementation generator.

Generators are specified by using a *_out flag. The argument given to this
type of flag can be either:
	1) *_out=some/directory/to/output/file(s)/to
	2) *_out=comma=separated,key=val,generator=option,pairs=then:some/directory/to/output/file(s)/to

An additional flag, *_opt, can be used to pass options to a generator. The
argument given to this type of flag is the same format as the *_opt
key=value pairs above.`,
	Example: "gqlc -I . --dart_out ./dartservice --doc_out ./docs --go_out ./goservice --js_out ./jsservice api.gql",
	Args:    cobra.MinimumNArgs(1), // Make sure at least one file is provided.
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
	gqlc flags files{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{$flags := in .LocalFlags "_out"}}{{if gt (len $flags.FlagUsages) 0}}

Generator Flags:
{{$flags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{$flags = ex .LocalFlags "_out"}}{{if gt (len $flags.FlagUsages) 0}}

General Flags:
{{$flags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Example:
	{{.Example}}{{end}}
`)
}

func runRoot(cmd *cobra.Command, args []string) (err error) {
	// Accumulate selected code generators
	var mode parser.Mode
	var gs []compiler.CodeGenerator
	flagOpts := make(map[compiler.CodeGenerator]string)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}

		g, isOpt := opts[f.Name]
		if isOpt {
			flagOpts[g] = f.Value.String()
			return
		}

		gen, exists := geners[f.Name]
		if !exists {
			return
		}
		gs = append(gs, gen)
	})

	// Parse files
	docs := make(map[string]*ast.Document, len(args))
	dset := token.NewDocSet()
	for _, filename := range args {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return err
		}

		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		doc, err := parser.ParseDoc(dset, filename, f, mode)
		if err != nil {
			return err
		}

		docs[doc.Name] = doc
	}

	// Perform type checking
	errs := util.CheckTypes(docs)
	if len(errs) > 0 {
		// TODO: Compound errors into a single error and return.
		return
	}

	// Run code generators
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, g := range gs {
		for _, doc := range docs {
			opt, _ := flagOpts[g]

			err = g.Generate(ctx, doc, opt)
			if err != nil {
				return
			}
		}
	}

	return
}
