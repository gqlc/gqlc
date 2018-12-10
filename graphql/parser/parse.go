// Package parser implements a parser for GraphQL IDL source files.
//
package parser

import (
	"fmt"
	"github.com/Zaba505/gqlc/graphql/ast"
	"github.com/Zaba505/gqlc/graphql/lexer"
	"github.com/Zaba505/gqlc/graphql/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

// Mode represents a parsing mode.
type Mode uint

const (
	ParseComments = 1 << iota // parse comments and add them to the schema
)

// ParseDir calls ParseDoc for all files with names ending in ".gql"/".graphql" in the
// directory specified by path and returns a map of document name -> *ast.Document for all
// the documents found.
//
func ParseDir(dset *token.DocSet, path string, filter func(os.FileInfo) bool, mode Mode) (docs map[string]*ast.Document, err error) {
	docs = make(map[string]*ast.Document)
	err = filepath.Walk(path, func(p string, info os.FileInfo, e error) error {
		skip := filter(info)
		if skip && info.IsDir() {
			return filepath.SkipDir
		}

		ext := filepath.Ext(p)
		if skip || info.IsDir() || ext != ".gql" && ext != ".graphql" {
			return nil
		}

		f, err := os.Open(info.Name())
		if err != nil {
			return err
		}

		doc, err := ParseDoc(dset, f.Name(), f, mode)
		f.Close() // TODO: Handle this error
		if err != nil {
			return err
		}

		docs[doc.Name] = doc
		return nil
	})
	return
}

// ParseDoc parses a single GraphQL Document.
//
func ParseDoc(dset *token.DocSet, name string, src io.Reader, mode Mode) (*ast.Document, error) {
	// Assume src isn't massive so we're gonna just read it all
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Create parser and doc to doc set. Then, parse doc.
	p := new(parser)
	d := dset.AddDoc(name, -1, len(b))
	return p.parse(d, b, mode)
}

// ParseDocs parses a set of GraphQL documents. Any import paths
// in a doc will be resolved against the provided doc names in the docs map.
//
func ParseDocs(dset *token.DocSet, docs map[string]io.Reader, mode Mode) (map[string]*ast.Document, error) {
	odocs := make(map[string]*ast.Document)

	for name, src := range docs {
		doc, err := ParseDoc(dset, name, src, mode)
		if err != nil {
			return odocs, err
		}
		odocs[name] = doc
	}
	return nil, nil
}

type parser struct {
	l    lexer.Interface
	name string
	line int
	pk   lexer.Item
	mode Mode
}

// next returns the next token
func (p *parser) next() (i lexer.Item) {
	if p.pk.Line != 0 {
		i = p.pk
		p.pk = lexer.Item{}
		return
	}
	return p.l.NextItem()
}

// peek peeks the next token
func (p *parser) peek() lexer.Item {
	p.pk = p.l.NextItem()
	return p.pk
}

// expect consumes the next token and guarantees it has the required type.
func (p *parser) expect(tok token.Token, context string) lexer.Item {
	i := p.l.NextItem()
	if i.Typ != tok {
		p.unexpected(i, context)
	}
	return i
}

// errorf formats the error and terminates processing.
func (p *parser) errorf(format string, args ...interface{}) {
	format = fmt.Sprintf("parser: %s:%d: %s", p.name, p.line, format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (p *parser) error(err error) {
	p.errorf("%s", err)
}

// unexpected complains about the token and terminates processing.
func (p *parser) unexpected(token lexer.Item, context string) {
	p.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of parse.
func (p *parser) recover(err *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if p != nil {
			p.l.Drain()
			p.l = nil
		}
		*err = e.(error)
	}
}

// ErrUnexpectedItem represents encountering an unexpected item from the lexer.
type ErrUnexpectedItem struct {
	i lexer.Item
}

// Error formats an ErrUnexpectedItem error.
func (e ErrUnexpectedItem) Error() string {
	return fmt.Sprintf("unexpected token encountered- line: %d, pos: %d, type: %s, value: %s", e.i.Line, e.i.Pos, e.i.Typ, e.i.String())
}

// parse parses a GraphQL document
func (p *parser) parse(doc *token.Doc, src []byte, mode Mode) (d *ast.Document, err error) {
	var lMode lexer.Mode
	if mode&ParseComments != 0 {
		lMode = lexer.ScanComments
	}

	p.l = lexer.Lex(doc, src, lMode)
	p.mode = mode

	d = &ast.Document{
		Name: doc.Name(),
		Doc:  new(ast.DocGroup),
	}
	defer p.recover(&err)
	p.parseDoc(d.Doc, d)
	return d, nil
}

// addDocs slurps up documentation
func (p *parser) addDocs(pdg *ast.DocGroup) (cdg *ast.DocGroup, item lexer.Item) {
	cdg = new(ast.DocGroup)
	for {
		// Get next item
		item = p.next()
		isComment := item.Typ == token.COMMENT
		if !isComment && item.Typ != token.DESCRIPTION {
			p.pk = lexer.Item{}
			return
		}

		// Skip a comment if they're not being parsed
		if isComment && p.mode&ParseComments == 0 {
			continue
		}
		cdg.List = append(cdg.List, &ast.Doc{
			Text:    item.Val,
			Char:    item.Pos,
			Comment: isComment,
		})

		// Peek next item.
		nitem := p.next()
		lineDiff := nitem.Line - item.Line
		if lineDiff == 1 {
			continue
		}

		// Add cdg to pdg
		pdg.List = append(pdg.List, cdg.List...)
	}
}

// parseDoc parses a GraphQL document
func (p *parser) parseDoc(dg *ast.DocGroup, d *ast.Document) {
	// Slurp up documentation
	cdg, item := p.addDocs(dg)

Loop:
	switch item.Typ {
	case token.ERR:
		p.unexpected(item, "parseDoc")
	case token.EOF:
		return
	case token.IMPORT:
		p.parseImport(item, cdg, d)
	case token.SCHEMA:
		p.parseSchema(item, cdg, d)
	case token.SCALAR:
		p.parseScalar(item, cdg, d)
	case token.TYPE:
		p.parseObject(item, cdg, d)
	case token.INTERFACE:
		p.parseInterface(item, cdg, d)
	case token.UNION:
		p.parseUnion(item, cdg, d)
	case token.ENUM:
		p.parseEnum(item, cdg, d)
	case token.INPUT:
		p.parseInput(item, cdg, d)
	case token.DIRECTIVE:
		p.parseDirective(item, cdg, d)
	case token.EXTEND:
		item = p.next()
		goto Loop
	}

	p.parseDoc(dg, d)
}

// parseImport parses a import declarations
func (p *parser) parseImport(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {
	// Create gen decl for import and add it the overall document
	imprtGen := &ast.GenDecl{
		Doc:    dg,
		TokPos: item.Pos,
		Tok:    token.IMPORT,
	}
	doc.Imports = append(doc.Imports, imprtGen)

	nitem := p.next()
	if nitem.Typ == token.LPAREN {
		imprtGen.Lparen = item.Pos
		nitem = p.next()
	}

	for {
		// Check for EOF or ERR
		if nitem.Typ == token.EOF || nitem.Typ == token.ERR {
			p.errorf("parser: unexpected token from lexer while parsing import: %s", nitem)
		}

		// Check for ')' in case of block import
		if nitem.Typ == token.RPAREN {
			imprtGen.Rparen = nitem.Pos
			break
		}

		// Check for comment
		if nitem.Typ == token.COMMENT && p.mode&ParseComments != 0 {
			imprtGen.Doc.List = append(imprtGen.Doc.List, &ast.Doc{
				Text:    nitem.Val,
				Char:    nitem.Pos,
				Comment: true,
			})
			nitem = p.next()
			continue
		}

		// Enforce strings only
		if nitem.Typ != token.STRING {
			p.unexpected(nitem, "parseImport")
		}

		// Create import spec node and add it to the larger import gen decl
		imprtSpec := &ast.ImportSpec{
			Name: &ast.Ident{},
			Path: &ast.BasicLit{
				ValuePos: nitem.Pos,
				Kind:     token.STRING,
				Value:    nitem.Val,
			},
		}
		imprtGen.Specs = append(imprtGen.Specs, imprtSpec)

		nitem = p.next()
	}
}

// parseSchema parses a schema declaration
func (p *parser) parseSchema(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {
	// Create schema general decl node
	schemaGen := &ast.GenDecl{
		Doc:    dg,
		Tok:    token.SCHEMA,
		TokPos: item.Pos,
	}
	doc.Schemas = append(doc.Schemas, schemaGen)
	doc.Types = append(doc.Types, schemaGen)

	// Slurp up applied directives
	dirs, nitem := p.parseDirectives()

	// Create schema type spec node
	schemaSpec := &ast.TypeSpec{
		Doc:  dg,
		Name: nil,
		Dirs: dirs,
	}
	schemaGen.Specs = append(schemaGen.Specs, schemaSpec)

	// Create schema type node
	schemaTyp := &ast.SchemaType{
		Schema: item.Pos,
		Fields: new(ast.FieldList),
	}
	schemaSpec.Type = schemaTyp

	if nitem.Typ != token.LBRACE {
		return
	}
	schemaTyp.Fields.Opening = nitem.Pos

	for {
		cdg, fitem := p.addDocs(dg)
		if fitem.Typ == token.RBRACE {
			schemaTyp.Fields.Closing = fitem.Pos
			return
		}

		if fitem.Typ != token.IDENT {
			p.unexpected(fitem, "parseSchema")
		}

		if fitem.Val != "query" && fitem.Val != "mutation" && fitem.Val != "subscription" {
			p.unexpected(fitem, "parseSchema")
		}

		f := &ast.Field{
			Doc: cdg,
			Name: &ast.Ident{
				NamePos: fitem.Pos,
				Name:    fitem.Val,
			},
		}
		schemaTyp.Fields.List = append(schemaTyp.Fields.List, f)

		p.expect(token.COLON, "parseSchema")

		fitem = p.expect(token.IDENT, "parseSchema")
		f.Type = &ast.Ident{
			NamePos: fitem.Pos,
			Name:    fitem.Val,
		}

		fitem = p.peek()
		if fitem.Typ == token.NOT {
			f.Type = &ast.NonNull{
				Type: f.Type,
			}
			p.pk = lexer.Item{}
		}
	}
}

// parseDirectives parses a list of applied directives
func (p *parser) parseDirectives() (dirs []ast.Expr, item lexer.Item) {
	item = p.next()
	for {
		if item.Typ != token.AT {
			return
		}
		dir := &ast.DirectiveLit{
			AtPos: item.Pos,
		}
		dirs = append(dirs, dir)

		item = p.expect(token.IDENT, "parseDirectives")
		dir.Name = item.Val

		item = p.next()
		if item.Typ != token.LPAREN {
			continue
		}

		args, rpos := p.parseArgs(nil) // TODO: Change nil to an actual *ast.DocGroup
		if args == nil {
			p.unexpected(p.pk, "parseDirectives:parseArgs")
		}

		dir.Args = &ast.CallExpr{
			Lparen: item.Pos,
			Args:   args,
			Rparen: rpos,
		}

		item = p.next()
	}
}

// parseArgs parses a list of arguments
func (p *parser) parseArgs(pdg *ast.DocGroup) (args []*ast.Arg, rpos token.Pos) {
	for {
		_, item := p.addDocs(pdg)
		if item.Typ == token.RPAREN {
			rpos = item.Pos
			return
		}

		if item.Typ != token.IDENT {
			p.unexpected(item, "parseArgs")
		}
		a := &ast.Arg{
			Name: &ast.Ident{
				NamePos: item.Pos,
				Name:    item.Val,
			},
		}
		args = append(args, a)

		p.expect(token.COLON, "parseArgsDef")

		a.Value, p.pk = p.parseValue()
	}
}

// TODO: parseValue parses a value.
func (p *parser) parseValue() (v ast.Expr, item lexer.Item) {
	return
}

// parseArgsDef parses a list of argument definitions.
func (p *parser) parseArgsDef(pdg *ast.DocGroup) (args []*ast.Field, rpos token.Pos) {
	for {
		cdg, item := p.addDocs(pdg)
		if item.Typ == token.RPAREN {
			rpos = item.Pos
			return
		}

		if item.Typ != token.IDENT {
			p.unexpected(item, "parseArgsDef")
		}
		f := &ast.Field{
			Doc: cdg,
			Name: &ast.Ident{
				NamePos: item.Pos,
				Name:    item.Val,
			},
		}
		args = append(args, f)

		p.expect(token.COLON, "parseArgsDef")

		item = p.expect(token.IDENT, "parseArgsDef")
		f.Type = &ast.Ident{
			NamePos: item.Pos,
			Name:    item.Val,
		}

		item = p.peek()
		if item.Typ == token.ASSIGN {
			item = p.expect(token.IDENT, "parseArgsDef")
			f.Default, item = p.parseValue()
		}

		f.Dirs, p.pk = p.parseDirectives()
	}
}

// parseScalar parses a scalar declaration
func (p *parser) parseScalar(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {
	scalarGen := &ast.GenDecl{
		Doc:    dg,
		TokPos: item.Pos,
		Tok:    token.SCALAR,
	}
	doc.Types = append(doc.Types, scalarGen)

	name := p.expect(token.IDENT, "parseScalar")

	scalarSpec := &ast.TypeSpec{
		Doc: dg,
		Name: &ast.Ident{
			NamePos: name.Pos,
			Name:    name.Val,
		},
	}

	scalarSpec.Dirs, item = p.parseDirectives()

	scalarType := &ast.ScalarType{
		Name: scalarSpec.Name,
	}
	scalarSpec.Type = scalarType
}

// TODO

// parseObject parses an object declaration
func (p *parser) parseObject(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}

// parseInterface parses an interface declaration
func (p *parser) parseInterface(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}

// parseUnion parses a union declaration
func (p *parser) parseUnion(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}

// parseEnum parses an enum declaration
func (p *parser) parseEnum(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}

// parseInput parses an input declaration
func (p *parser) parseInput(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}

// parseDirective parses a directive declaration
func (p *parser) parseDirective(item lexer.Item, dg *ast.DocGroup, doc *ast.Document) {}