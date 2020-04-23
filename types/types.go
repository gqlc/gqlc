// Package types just provides a wrapper around the internal
// compiler type registration. This is used to distinguish
// gqlc builtin types from compiler builtin types.
//
package types

import (
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

var dirs []string

// Register registers the types with compiler and tracks custom directives.
func Register(decls ...*ast.TypeDecl) {
	for _, decl := range decls {
		if decl.Tok != token.Token_DIRECTIVE {
			continue
		}

		dirs = append(dirs, decl.Spec.(*ast.TypeDecl_TypeSpec).TypeSpec.Name.Name)
	}

	compiler.RegisterTypes(decls...)
}

// IsGqlcDirective checks if a directive is custom to gqlc or not.
func IsGqlcDirective(name string) bool {
	for _, d := range dirs {
		if d == name {
			return true
		}
	}
	return false
}
