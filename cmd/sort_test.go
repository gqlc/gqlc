package cmd

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/gqlc/graphql/ast"
)

func TestSortTypeDecls(t *testing.T) {
	testCases := []struct {
		name   string
		types  []*ast.TypeSpec
		sorted []*ast.TypeSpec
	}{
		{
			name: "ByName",
			types: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
			},
			sorted: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
			},
		},
		{
			name: "ByType",
			types: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Schema{
						Schema: &ast.SchemaType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Enum{
						Enum: &ast.EnumType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Object{
						Object: &ast.ObjectType{},
					},
				},
			},
			sorted: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Schema{
						Schema: &ast.SchemaType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Object{
						Object: &ast.ObjectType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Enum{
						Enum: &ast.EnumType{},
					},
				},
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
			},
		},
		{
			name: "ByBoth",
			types: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
			},
			sorted: []*ast.TypeSpec{
				{
					Name: &ast.Ident{Name: "a"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "q"},
					Type: &ast.TypeSpec_Scalar{
						Scalar: &ast.ScalarType{},
					},
				},
				{
					Name: &ast.Ident{Name: "n"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
				{
					Name: &ast.Ident{Name: "z"},
					Type: &ast.TypeSpec_Directive{
						Directive: &ast.DirectiveType{},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(subT *testing.T) {
			types := make([]*ast.TypeDecl, 0, len(testCase.types))
			for _, spec := range testCase.types {
				types = append(types, &ast.TypeDecl{Spec: &ast.TypeDecl_TypeSpec{TypeSpec: spec}})
			}

			stypes := sortTypeDecls(types)

			for i, st := range stypes {
				cmpSpec := st.Spec.(*ast.TypeDecl_TypeSpec).TypeSpec

				ogb, err := proto.Marshal(testCase.sorted[i])
				if err != nil {
					subT.Error(err)
					return
				}

				cb, err := proto.Marshal(cmpSpec)
				if err != nil {
					subT.Error(err)
					return
				}

				if !bytes.Equal(cb, ogb) {
					subT.Fail()
				}
			}
		})
	}
}
