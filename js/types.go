// types.go contains the GraphQL types this generator supports

package js

import (
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

var types = []*ast.TypeDecl{
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "js"},
			Type: &ast.TypeSpec_Directive{Directive: &ast.DirectiveType{
				Locs: []*ast.DirectiveLocation{{Loc: ast.DirectiveLocation_DOCUMENT}},
				Args: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "options"},
							Type: &ast.InputValue_Ident{
								Ident: &ast.Ident{Name: "JsOptions"},
							},
						},
					},
				},
			}},
		}},
	},
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "JsOptions"},
			Type: &ast.TypeSpec_Input{Input: &ast.InputType{
				Fields: &ast.InputValueList{
					List: []*ast.InputValue{
						{
							Name: &ast.Ident{Name: "module"},
							Type: &ast.InputValue_NonNull{NonNull: &ast.NonNull{
								Type: &ast.NonNull_Ident{
									Ident: &ast.Ident{
										Name: "Module",
									},
								},
							}},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_IDENT,
								Value: "COMMONJS",
							}},
						},
						{
							Name: &ast.Ident{Name: "useFlow"},
							Type: &ast.InputValue_Ident{
								Ident: &ast.Ident{Name: "Boolean"},
							},
							Default: &ast.InputValue_BasicLit{BasicLit: &ast.BasicLit{
								Kind:  token.Token_BOOL,
								Value: "false",
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
	{
		Spec: &ast.TypeDecl_TypeSpec{TypeSpec: &ast.TypeSpec{
			Name: &ast.Ident{Name: "Module"},
			Type: &ast.TypeSpec_Enum{Enum: &ast.EnumType{
				Values: &ast.FieldList{
					List: []*ast.Field{
						{
							Name: &ast.Ident{Name: "COMMONJS"},
						},
						{
							Name: &ast.Ident{Name: "ES6"},
						},
					},
				},
			}},
		}},
	},
}

func init() {
	compiler.RegisterTypes(types...)
}
