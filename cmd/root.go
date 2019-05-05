package cmd

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
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
	Example:            "gqlc -I . --dart_out ./dartservice --doc_out ./docs --go_out ./goservice --js_out ./jsservice api.gql",
	DisableFlagParsing: true,
	Args:               cobra.MinimumNArgs(1), // Make sure at least one file is provided.
}

func init() {
	cobra.AddTemplateFuncs(tmplFs)

	rootCmd.Flags().StringSliceP("import_path", "I", []string{"."}, `Specify the directory in which to search for
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

func preRunRoot(geners, opts map[string]compiler.Generator, genOpts map[compiler.Generator]*oFlag) func(cmd *cobra.Command, args []string) error {
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
			pg := &pluginGenerator{Name: strings.TrimSuffix(name, "_out")}

			outFlag := *f
			outFlag.isOut = true
			cmd.Flags().Var(outFlag, name, "")
			geners[a] = pg

			optName := strings.Replace(name, "_out", "_opt", 1)
			optFlag := *f
			cmd.Flags().Var(optFlag, optName, "")
			opts[optName] = pg
		}
		if err := cmd.Flags().Parse(args); err != nil {
			return err
		}
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

		// Accumulate selected code generators
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

type genCtx struct {
	fs  afero.Fs
	dir string
}

func (ctx *genCtx) Open(name string) (io.WriteCloser, error) {
	return ctx.fs.OpenFile(filepath.Join(ctx.dir, name), os.O_WRONLY|os.O_CREATE, 0755)
}

func runRoot(fs afero.Fs, genOpts map[compiler.Generator]*oFlag) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		importPaths, err := cmd.Flags().GetStringSlice("import_path")
		if err != nil {
			return
		}

		// Parse files
		docs, err := parseInputFiles(fs, importPaths, cmd.Flags().Args())
		if err != nil {
			return
		}

		// Perform type checking
		errs := compiler.CheckTypes(docs)
		if len(errs) > 0 {
			// TODO: Compound errors into a single error and return.
			return
		}

		// Resolve imports
		docs, err = compiler.ReduceImports(docs)
		if err != nil {
			return
		}

		// Run code generators
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var b bytes.Buffer
		enc := json.NewEncoder(&b)
		for g, genOpt := range genOpts {
			err = enc.Encode(genOpt.opts)
			if err != nil {
				return
			}

			ctx = compiler.WithContext(ctx, &genCtx{dir: *genOpt.outDir, fs: fs})

			for _, doc := range docs {
				err = g.Generate(ctx, doc, b.String())
				if err != nil {
					return
				}
			}
			b.Reset()
		}

		return
	}
}

// resolveImportPaths makes sure import paths and doc names are consistent.
func resolveImportPaths(docs []*ast.Document) {
	for _, d := range docs {
		filename := filepath.Base(d.Name)
		d.Name = filename[:len(filename)-len(filepath.Ext(filename))]

		for _, direc := range d.Directives {
			if direc.Name != "import" {
				continue
			}

			for _, arg := range direc.Args.Args {

				compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
				listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List
				paths := listLit.(*ast.ListLit_BasicList).BasicList

				for _, p := range paths.Values {
					iPath := strings.Trim(p.Value, "\"")
					iName := filepath.Base(iPath)

					p.Value = fmt.Sprintf(`"%s"`, iName[:len(iName)-len(filepath.Ext(iName))])
				}
			}
		}
	}
}

// parseInputFiles parses all input files from the command line args, as well as any imported files.
func parseInputFiles(fs afero.Fs, importPaths []string, args []string) (docs []*ast.Document, err error) {
	// After successful parsing, import path resolution must be
	// done to prevent any issues when doing later optimizations e.g.
	// type checking, import resolution, etc.
	defer func() {
		if err == nil {
			resolveImportPaths(docs)
		}
	}()

	// Parse files from args
	docMap := make(map[string]*ast.Document, len(args))
	dset := token.NewDocSet()
	for _, filename := range args {
		f, err := openFile(fs, importPaths, filename)
		if err != nil {
			return nil, err
		}

		name := filepath.Base(filename)
		doc, err := parser.ParseDoc(dset, name, f, 0)
		if err != nil {
			return nil, err
		}

		docMap[name] = doc
	}

	// Parse imports
	err = parseImports(dset, fs, importPaths, docMap)
	if err != nil {
		return
	}

	// Convert docMap to doc slice
	docs = make([]*ast.Document, 0, len(docMap))
	for _, doc := range docMap {
		docs = append(docs, doc)
	}

	return
}

func parseImports(dset *token.DocSet, fs afero.Fs, importPaths []string, docMap map[string]*ast.Document) error {
	type iInfo struct {
		Name string
		Path string
	}
	q := list.New()
	for _, doc := range docMap {
		for _, direc := range doc.Directives {
			if direc.Name != "import" {
				continue
			}

			for _, arg := range direc.Args.Args {

				compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
				listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List
				paths := listLit.(*ast.ListLit_BasicList).BasicList

				for _, p := range paths.Values {
					iPath := strings.Trim(p.Value, "\"")
					iName := filepath.Base(iPath)
					if _, exists := docMap[iName]; exists {
						continue
					}

					q.PushBack(iInfo{Name: iName, Path: iPath})
				}
			}
		}
	}
	for q.Len() > 0 {
		e := q.Front()
		q.Remove(e)
		i := e.Value.(iInfo)

		f, err := openFile(fs, importPaths, i.Path)
		if err != nil {
			return err
		}

		d, err := parser.ParseDoc(dset, i.Name, f, 0)
		if err != nil {
			return err
		}

		docMap[i.Name] = d

		for _, direc := range d.Directives {
			if direc.Name != "import" {
				continue
			}

			for _, arg := range direc.Args.Args {

				compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
				listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List
				paths := listLit.(*ast.ListLit_BasicList).BasicList

				for _, p := range paths.Values {
					iPath := strings.Trim(p.Value, "\"")
					iName := filepath.Base(iPath)
					if _, exists := docMap[iName]; exists {
						continue
					}

					q.PushBack(iInfo{Name: iName, Path: iPath})
				}
			}
		}
	}

	return nil
}

// openFile is just a helper for opening files
func openFile(fs afero.Fs, importPaths []string, filename string) (f afero.File, err error) {
	// Check if filename if Abs
	var exists bool
	if !filepath.IsAbs(filename) {
		for _, iPath := range importPaths {
			fname := filepath.Join(iPath, filename)
			exists, err = afero.Exists(fs, fname)
			if err != nil {
				return
			}

			if exists {
				filename = fname
				break
			}
		}
	}

	f, err = fs.Open(filename)
	return
}
