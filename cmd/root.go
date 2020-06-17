package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/scanner"
	"time"

	"github.com/gqlc/compiler"
	"github.com/gqlc/compiler/spec"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type gqlcCmd struct {
	*cobra.Command
}

func (c *CommandLine) newGqlcCmd(cfgs []genConfig, fs afero.Fs, pluginPrefix string) *gqlcCmd {
	outDirs := make([]string, 0, len(cfgs))
	geners := make([]generator, 0, len(cfgs))

	restoreLogger := func() {}

	cmd := &gqlcCmd{
		Command: &cobra.Command{
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
			Example: "gqlc -I . --doc_out ./docs --go_out ./goservice --js_out ./jsservice api.gql",
			Args: func(cmd *cobra.Command, args []string) error {
				err := cobra.MinimumNArgs(1)(cmd, args)
				if err != nil {
					return err
				}

				return validateFilenames(cmd, args)
			},
			PreRunE: chainPreRunEs(
				func(cmd *cobra.Command, args []string) error {
					v, err := cmd.Flags().GetBool("verbose")
					if !v || err != nil {
						return err
					}

					enc := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
					core := zapcore.NewCore(enc, os.Stdout, zap.InfoLevel)

					l := zap.New(core)
					restoreLogger = zap.ReplaceGlobals(l)
					return err
				},
				validatePluginTypes(c.fs),
				initGenDirs(fs, &outDirs),
			),
			RunE: func(cmd *cobra.Command, args []string) error {
				defer restoreLogger()
				defer zap.L().Sync()

				importPaths, err := cmd.Flags().GetStringSlice("import_path")
				if err != nil {
					return err
				}

				return root(fs, geners, importPaths, cmd.Flags().Args()...)
			},
			SilenceUsage:  true,
			SilenceErrors: true,
		},
	}

	cmd.SetUsageTemplate(usageTmpl)

	cmd.Flags().StringSliceP("import_path", "I", []string{"."}, `Specify the directory in which to search for
imports.  May be specified multiple times;
directories will be searched in order.  If not
given, the current working directory is used.`)
	cmd.Flags().BoolP("verbose", "v", false, "Output logging")
	cmd.Flags().StringSliceP("types", "t", nil, "Provide .gql files containing types you wish to register with the compiler.")

	fp := &fparser{
		Scanner: new(scanner.Scanner),
	}

	for _, cfg := range cfgs {
		f := genFlag{
			g:       cfg.g,
			opts:    make(map[string]interface{}),
			geners:  &geners,
			outDirs: &outDirs,
			fp:      fp,
		}

		cmd.Flags().Var(f, cfg.name, cfg.help)

		if cfg.opt != "" {
			f.isOpt = true
			cmd.Flags().Var(f, cfg.opt, "Pass additional options to generator.")
		}
	}

	return cmd
}

type genCtx struct {
	fs  afero.Fs
	dir string
}

func (ctx *genCtx) Open(name string) (io.WriteCloser, error) {
	f, err := ctx.fs.OpenFile(filepath.Join(ctx.dir, name), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	return f, f.Truncate(0)
}

type generator struct {
	gen.Generator

	opts   map[string]interface{}
	outDir string
}

func root(fs afero.Fs, geners []generator, iPaths []string, args ...string) (err error) {
	// Parse files
	zap.S().Info("parsing input files")
	docMap := make(map[string]*ast.Document, len(args))
	err = parseInputFiles(fs, token.NewDocSet(), docMap, iPaths, args...)
	if err != nil {
		return
	}

	zap.S().Info("resolving import paths")
	docs := make([]*ast.Document, 0, len(docMap))
	for _, doc := range docMap {
		docs = append(docs, doc)
	}
	resolveImportPaths(docs)

	docsIR := compiler.ToIR(docs)

	// Resolve imports (this must occur before type checking)
	zap.S().Info("reducing imports")
	docsIR, err = compiler.ReduceImports(docsIR)
	if err != nil {
		return err
	}

	// Add any missing fields to objects that implement interfaces
	zap.S().Info("implementing interfaces")
	docsIR = implInterfaces(docsIR)

	// Perform type checking
	zap.S().Info("type checking")
	errs := compiler.CheckTypes(docsIR, spec.Validator, compiler.ImportValidator)
	if len(errs) > 0 {
		for _, err = range errs {
			log.Println(err)
		}
		return
	}

	// Merge type extensions with the original type definitions
	zap.S().Info("merging type extensions")
	for d, types := range docsIR {
		docsIR[d] = compiler.MergeExtensions(types)
	}

	// Convert types from IR to []*ast.TypeDecl
	docs = compiler.FromIR(docsIR)
	for _, doc := range docs {
		doc.Types = sortTypeDecls(doc.Types)
	}

	// Run code generators
	zap.S().Info("generating documents")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, g := range geners {
		ctx = gen.WithContext(ctx, &genCtx{dir: g.outDir, fs: fs})

		for _, doc := range docs {
			err = g.Generate(ctx, doc, g.opts)
			if err != nil {
				return
			}
		}
	}
	return
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

				var paths []*ast.BasicLit
				switch v := listLit.(type) {
				case *ast.ListLit_BasicList:
					paths = append(paths, v.BasicList.Values...)
				case *ast.ListLit_CompositeList:
					cpaths := v.CompositeList.Values
					paths = make([]*ast.BasicLit, len(cpaths))
					for i, c := range cpaths {
						paths[i] = c.Value.(*ast.CompositeLit_BasicLit).BasicLit
					}
				}

				for _, p := range paths {
					iPath := strings.Trim(p.Value, "\"")
					iName := filepath.Base(iPath)

					p.Value = fmt.Sprintf(`"%s"`, iName[:len(iName)-len(filepath.Ext(iName))])
				}
			}
		}
	}
}

// parseInputFiles parses all input files from the command line args, as well as any imported files.
func parseInputFiles(fs afero.Fs, dset *token.DocSet, docs map[string]*ast.Document, importPaths []string, filenames ...string) error {
	for _, filename := range filenames {
		name := filepath.Base(filename)
		if _, exists := docs[name]; exists {
			continue
		}

		f, err := openFile(filename, fs, importPaths)
		if err != nil {
			return err
		}
		defer f.Close()

		doc, err := parser.ParseDoc(dset, name, f, parser.ParseComments)
		if err != nil {
			return err
		}

		docs[name] = doc
	}

	for _, doc := range docs {
		imports := getImports(doc)
		imports = filter(imports, docs, fs, importPaths)
		if len(imports) == 0 {
			continue
		}

		err := parseInputFiles(fs, dset, docs, importPaths, imports...)
		if err != nil {
			return err
		}
	}

	return nil
}

func getImports(doc *ast.Document) (names []string) {
	for _, direc := range doc.Directives {
		if direc.Name != "import" {
			continue
		}

		for _, arg := range direc.Args.Args {

			compLit := arg.Value.(*ast.Arg_CompositeLit).CompositeLit
			listLit := compLit.Value.(*ast.CompositeLit_ListLit).ListLit.List

			var paths []*ast.BasicLit
			switch v := listLit.(type) {
			case *ast.ListLit_BasicList:
				paths = append(paths, v.BasicList.Values...)
			case *ast.ListLit_CompositeList:
				cpaths := v.CompositeList.Values
				paths = make([]*ast.BasicLit, len(cpaths))
				for i, c := range cpaths {
					paths[i] = c.Value.(*ast.CompositeLit_BasicLit).BasicLit
				}
			}

			for _, p := range paths {
				names = append(names, strings.Trim(p.Value, "\""))
			}
		}
	}

	return
}

var client = &fetchClient{
	Client: &http.Client{
		Timeout: 5 * time.Second,
	},
}

func openFile(name string, fs afero.Fs, importPaths []string) (io.ReadCloser, error) {
	endpoint, err := url.Parse(name)
	if err != nil {
		return nil, err
	}
	if endpoint.Scheme != "" && endpoint.Opaque == "" {
		return fetch(client, endpoint)
	}

	fname, err := normFilePath(fs, importPaths, name)
	if err != nil {
		return nil, err
	}
	if fname == "" {
		return nil, fmt.Errorf("could not resolve file path: %s", name)
	}

	return fs.Open(fname)
}

// normFilePath converts any path to absolute path
func normFilePath(fs afero.Fs, iPaths []string, filename string) (string, error) {
	if filepath.IsAbs(filename) {
		return filename, nil
	}

	for _, iPath := range iPaths {
		fname := filepath.Join(iPath, filename)
		exists, err := afero.Exists(fs, fname)
		if err != nil {
			return "", err
		}

		if exists {
			return fname, nil
		}
	}
	return "", nil
}

// filter filters the strings in b from a
func filter(a []string, b map[string]*ast.Document, fs afero.Fs, iPaths []string) []string {
	n := 0
	for _, x := range a {
		name := filepath.Base(x)

		if _, exists := b[name]; !exists {
			a[n] = x
			n++
		}
	}
	return a[:n]
}
