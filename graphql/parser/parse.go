// Package parser implements a parser for GraphQL IDL source files.
//
package parser

import (
	"fmt"
	"gqlc/graphql/ast"
	"gqlc/graphql/lexer"
	"gqlc/graphql/token"
	"io"
	"io/ioutil"
	"os"
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
