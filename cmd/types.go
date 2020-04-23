// types.go contains builtin types for the compiler

package cmd

import (
	"github.com/gqlc/gqlc/types"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

var gqlcTypes = []*ast.TypeDecl{
	{
		Tok: token.Token_DIRECTIVE,
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "resolver"},
			Type: &ast.TypeSpec_Directive{Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{
					{Loc: ast.DirectiveLocation_FIELD_DEFINITION},
					{Loc: ast.DirectiveLocation_UNION},
				},
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
	{
		Tok: token.Token_DIRECTIVE,
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "as"},
			Type: &ast.TypeSpec_Directive{Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{
					{Loc: ast.DirectiveLocation_ENUM_VALUE},
				},
				Args: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "value"},
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
	types.Register(gqlcTypes...)
}
