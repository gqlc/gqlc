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

type gqlcConfig struct {
	ipaths []string
	geners []generator

	logger  *zap.Logger
	client  *fetchClient
	headers http.Header
}

type gqlcCmd struct {
	*cobra.Command

	cfg *gqlcConfig
}

func (c *CommandLine) newGqlcCmd(cfgs []genConfig, fs afero.Fs, pluginPrefix string) *gqlcCmd {
	outDirs := make([]string, 0, len(cfgs))

	cc := &gqlcCmd{
		cfg: &gqlcConfig{
			geners:  make([]generator, 0, len(cfgs)),
			client:  defaultClient,
			headers: make(http.Header),
		},
	}

	cc.Command = &cobra.Command{
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

				cc.cfg.logger = zap.New(core)
				return err
			},
			func(cmd *cobra.Command, args []string) (err error) {
				cc.cfg.ipaths, err = cmd.Flags().GetStringSlice("import_path")
				return
			},
			cc.validatePluginTypes(c.fs),
			initGenDirs(fs, &outDirs),
		),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if cc.cfg.logger == nil {
					return
				}
				zap.ReplaceGlobals(cc.cfg.logger)
			}()
			defer zap.L().Sync()

			return cc.run(fs, cmd.Flags().Args()...)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cc.SetUsageTemplate(usageTmpl)

	cc.Flags().StringSliceP("import_path", "I", []string{"."}, `Specify the directory in which to search for
imports.  May be specified multiple times;
directories will be searched in order.  If not
given, the current working directory is used.`)
	cc.Flags().BoolP("verbose", "v", false, "Output logging")
	cc.Flags().StringSliceP("types", "t", nil, "Provide .gql files containing types you wish to register with the compiler.")
	cc.Flags().VarP(&headerFlag{value: &cc.cfg.headers}, "headers", "H", "Provide HTTP headers to fetching. Format: a=1,b=2")

	fp := &fparser{
		Scanner: new(scanner.Scanner),
	}

	for _, cfg := range cfgs {
		f := genFlag{
			g:       cfg.g,
			opts:    make(map[string]interface{}),
			geners:  &cc.cfg.geners,
			outDirs: &outDirs,
			fp:      fp,
		}

		cc.Flags().Var(f, cfg.name, cfg.help)

		if cfg.opt != "" {
			f.isOpt = true
			cc.Flags().Var(f, cfg.opt, "Pass additional options to generator.")
		}
	}

	return cc
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

func (c *gqlcCmd) run(fs afero.Fs, args ...string) (err error) {
	// Parse files
	zap.S().Info("parsing input files")
	docMap := make(map[string]*ast.Document, len(args))
	err = c.parseInputFiles(fs, token.NewDocSet(), docMap, args...)
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
	for _, g := range c.cfg.geners {
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
func (c *gqlcCmd) parseInputFiles(fs afero.Fs, dset *token.DocSet, docs map[string]*ast.Document, filenames ...string) error {
	for _, filename := range filenames {
		name := filepath.Base(filename)
		if _, exists := docs[name]; exists {
			continue
		}

		f, err := c.openFile(filename, fs)
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
		imports = filter(imports, docs, fs, c.cfg.ipaths)
		if len(imports) == 0 {
			continue
		}

		err := c.parseInputFiles(fs, dset, docs, imports...)
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

func (c *gqlcCmd) openFile(name string, fs afero.Fs) (io.ReadCloser, error) {
	endpoint, err := url.Parse(name)
	if err != nil {
		return nil, err
	}
	if endpoint.Scheme != "" && endpoint.Opaque == "" {
		return fetch(c.cfg.client, endpoint, c.cfg.headers)
	}

	fname, err := normFilePath(fs, c.cfg.ipaths, name)
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
