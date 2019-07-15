// types.go contains builtin types for the compiler

package cmd

import (
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
)

var gqlcTypes = []*ast.TypeDecl{
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "resolver"},
			Type: &ast.TypeSpec_Directive{Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{{Loc: ast.DirectiveLocation_FIELD_DEFINITION}},
				Args: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "name"},
							Type: &ast.InputValue_Ident{
								Ident: &ast.Ident{Name: "String"},
							},
						},
					},
				},
			}},
		}},
	},
}

func init() {
	compiler.RegisterTypes(gqlcTypes...)
}
