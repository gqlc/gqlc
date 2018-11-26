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
	"runtime"
)

type Mode uint

const (
	ParseComments = 1 << iota // parse comments and add them to the schema
)

// ParseDir calls ParseFile for all files with names ending in ".gql"/".graphql" in the
// directory specified by path and returns a map of file name -> File Schema with all
// the schemas found.
func ParseDir(fset *token.FileSet, path string, filter func(os.FileInfo) bool, mode Mode) (map[string]*ast.Document, error) {
	return nil, nil
}

// ParseFile parses a single GraphQL Schema file.
func ParseFile(fset *token.FileSet, filename string, src io.Reader, mode Mode) (*ast.Document, error) {
	// Assume src isn't massive so we're gonna just read it all
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	var m lexer.Mode
	if mode == ParseComments {
		m = lexer.ScanComments
	}
	l := lexer.Lex(fset.AddFile(filename, -1, len(b)), b, m)
	i := lexer.Item{}
	for item := l.NextItem(); ; {
		fmt.Println(item)
		item = l.NextItem()
		if item == i {
			break
		}
	}
	return nil, nil
}

// ParseDocument parses a GraphQL document read from the provided reader.
// It makes no assumption of the origin of the src, thus allowing it to be
// used a little bit more freely than ParseFile.
func ParseDocument(src io.Reader, mode Mode) (*ast.Document, error) {
	return nil, nil
}

type parser struct {
	l    lexer.Interface
	name string
	line int
	pk   lexer.Item
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
	}
}

// ErrUnexpectedItem represents encountering an unexpected item from the lexer.
type ErrUnexpectedItem struct {
	i lexer.Item
}

func (e ErrUnexpectedItem) Error() string {
	return fmt.Sprintf("unexpected token encountered- line: %d, pos: %d, type: %s, value: %s", e.i.Line, e.i.Pos, e.i.Typ, e.i.String())
}

func (p *parser) parse(file *token.File, src []byte, mode Mode) (doc *ast.Document, err error) {
	var lMode lexer.Mode
	if mode&ParseComments != 0 {
		lMode = lexer.ScanComments
	}

	p.l = lexer.Lex(file, src, lMode)

	doc = new(ast.Document)
	defer p.recover(&err)
	for item := p.l.NextItem(); item.Typ != token.ERR && item.Typ != token.EOF; {
		switch item.Typ {
		case token.COMMENT, token.DESCRIPTION:
			// TODO
		case token.IMPORT:
			p.parseImport(doc)
		case token.SCHEMA:
			p.parseSchema(doc)
		case token.SCALAR:
			p.parseScalar(doc)
		case token.TYPE:
			p.parseObject(doc)
		case token.INTERFACE:
			p.parseInterface(doc)
		case token.UNION:
			p.parseUnion(doc)
		case token.ENUM:
			p.parseEnum(doc)
		case token.INPUT:
			p.parseInput(doc)
		case token.DIRECTIVE:
			p.parseDirective(doc)
		default:
			return nil, ErrUnexpectedItem{item}
		}
		item = p.l.NextItem()
	}
	return doc, nil
}

func (p *parser) parseImport(doc *ast.Document) {}

func (p *parser) parseSchema(doc *ast.Document) {}

func (p *parser) parseScalar(doc *ast.Document) {}

func (p *parser) parseObject(doc *ast.Document) {}

func (p *parser) parseInterface(doc *ast.Document) {}

func (p *parser) parseUnion(doc *ast.Document) {}

func (p *parser) parseEnum(doc *ast.Document) {}

func (p *parser) parseInput(doc *ast.Document) {}

func (p *parser) parseDirective(doc *ast.Document) {}
