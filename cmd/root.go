package cmd

import (
	"bytes"
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gqlc/compiler"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/parser"
	"github.com/gqlc/graphql/token"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
	Example:            "gqlc -I . --doc_out ./docs --go_out ./goservice --js_out ./jsservice api.gql",
	DisableFlagParsing: true,
}

func init() {
	rootCmd.Flags().StringSliceP("import_path", "I", []string{"."}, `Specify the directory in which to search for
imports.  May be specified multiple times;
directories will be searched in order.  If not
given, the current working directory is used.`)
	rootCmd.Flags().BoolP("verbose", "v", false, "Output logging")
	rootCmd.SetUsageTemplate(usageTmpl)
}

type genCtx struct {
	fs  afero.Fs
	dir string
}

func (ctx *genCtx) Open(name string) (io.WriteCloser, error) {
	return ctx.fs.OpenFile(filepath.Join(ctx.dir, name), os.O_WRONLY|os.O_CREATE, 0755)
}

func root(fs afero.Fs, geners *[]*genFlag, iPaths []string, args ...string) (err error) {
	// Parse files
	docMap := make(map[string]*ast.Document, len(args))
	err = parseInputFiles2(fs, docMap, iPaths, args...)
	if err != nil {
		return
	}

	docs := make([]*ast.Document, 0, len(docMap))
	for _, doc := range docMap {
		docs = append(docs, doc)
	}
	resolveImportPaths(docs)

	// First, Resolve imports (this must occur before type checking)
	docsIR, err := compiler.ReduceImports(docs)
	if err != nil {
		return err
	}

	// Then, Perform type checking
	errs := compiler.CheckTypes(docsIR, compiler.TypeCheckerFn(compiler.Validate))
	if len(errs) > 0 {
		for _, err = range errs {
			log.Println(err)
		}
		return
	}

	// Merge type extensions with the original type definitions
	for d, types := range docsIR {
		docsIR[d] = compiler.MergeExtensions(types)
	}

	// Remove builtin types and any Generator registered types
	builtins := compiler.ToIR(compiler.Types)
	for name := range builtins {
		for _, d := range docsIR {
			delete(d, name)
		}
	}

	// Convert types from IR to []*ast.TypeDecl
	docs = docs[:0]
	for doc, types := range docsIR {
		doc.Types = compiler.FromIR(types)

		docs = append(docs, doc)
	}

	// Run code generators
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	for _, g := range *geners {
		b.Reset()

		err = enc.Encode(g.opts)
		if err != nil {
			return
		}

		ctx = gen.WithContext(ctx, &genCtx{dir: *g.outDir, fs: fs})

		for _, doc := range docs {
			err = g.Generate(ctx, doc, b.String())
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
		doc, err := parser.ParseDoc(dset, name, f, parser.ParseComments)
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

func parseInputFiles2(fs afero.Fs, docs map[string]*ast.Document, importPaths []string, filenames ...string) error {
	dset := token.NewDocSet()

	for _, filename := range filenames {
		name := filepath.Base(filename)
		if _, exists := docs[name]; exists {
			continue
		}

		f, err := openFile(fs, importPaths, filename)
		if err != nil {
			return err
		}

		doc, err := parser.ParseDoc(dset, name, f, parser.ParseComments)
		if err != nil {
			return err
		}

		docs[name] = doc

		imports := getImports2(doc)
		err = parseInputFiles2(fs, docs, importPaths, imports...)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseImports(dset *token.DocSet, fs afero.Fs, importPaths []string, docMap map[string]*ast.Document) error {
	q := list.New()
	for _, doc := range docMap {
		getImports(doc, q, docMap)
	}

	for q.Len() > 0 {
		e := q.Front()
		q.Remove(e)
		i := e.Value.(iInfo)

		if _, exists := docMap[i.Name]; exists {
			continue
		}

		f, err := openFile(fs, importPaths, i.Path)
		if err != nil {
			return err
		}

		d, err := parser.ParseDoc(dset, i.Name, f, 0)
		if err != nil {
			return err
		}

		docMap[i.Name] = d

		getImports(d, q, docMap)
	}

	return nil
}

type iInfo struct {
	Name string
	Path string
}

func getImports(doc *ast.Document, q *list.List, docMap map[string]*ast.Document) {
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

func getImports2(doc *ast.Document) (names []string) {
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
