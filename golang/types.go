package golang

import (
	"github.com/gqlc/gqlc/types"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

var goTypes = []*ast.TypeDecl{
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "go"},
			Type: &ast.TypeSpec_Directive{Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{{Loc: ast.DirectiveLocation_DOCUMENT}},
				Args: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "options"},
							Type: &ast.InputValue_Ident{
								Ident: &ast.Ident{Name: "GoOptions"},
							},
						},
					},
				},
			}},
		}},
	},
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "GoOptions"},
			Type: &ast.TypeSpec_Input{Input: &ast.InputType{
				Fields: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "package"},
							Type: &ast.InputValue_NonNull{NonNull: &ast.NonNull{
								Type: &ast.NonNull_Ident{
									Ident: &ast.Ident{
										Name: "String",
									},
								},
							}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_STRING,
								Value: `"main"`,
							}},
						},
						{
							Name: &ast.Ident{Name: "descriptions"},
							Type: &ast.InputValue_Ident{
								Ident: &ast.Ident{Name: "Boolean"},
							},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_BOOL,
								Value: "false",
							}},
						},
					},
				},
			}},
		}},
	},
}

func init() {
	types.Register(goTypes...)
}
