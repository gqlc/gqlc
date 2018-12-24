package util

import "github.com/gqlc/graphql/ast"

// TypeError represents a type error.
type TypeError struct {
	// Document where type error was discovered
	Doc *ast.Doc

	// Type error message
	Msg string
}

// Error returns a string representation of a TypeError.
func (e *TypeError) Error() string {
	return e.Msg
}

// CheckTypes type checks a set of GraphQL documents.
// All type errors will be collected. Only one schema
// is allowed in a set of Documents.
//
func CheckTypes(docs map[string]*ast.Document) []*TypeError {
	// TODO
	return nil
}
