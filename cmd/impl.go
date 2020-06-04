// impl.go adds missings fields to objects the implement interfaces

package cmd

import (
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

func implInterfaces(ir compiler.IR) compiler.IR {
	for _, types := range ir {
		for _, decls := range types {
			for _, decl := range decls {
				if decl.Tok != token.Token_TYPE {
					continue
				}

				s := decl.Spec.(*ast.TypeDecl_TypeSpec).TypeSpec
				obj := s.Type.(*ast.TypeSpec_Object).Object
				if len(obj.Interfaces) == 0 {
					continue
				}

				applyInterfaces(ir, obj.Fields, obj.Interfaces)
			}
		}
	}
	return ir
}

func applyInterfaces(ir compiler.IR, objFields *ast.FieldList, interfaces []*ast.Ident) {
	for _, id := range interfaces {
		_, idecls := compiler.Lookup(id.Name, ir)
		is := idecls[0].Spec.(*ast.TypeDecl_TypeSpec).TypeSpec
		it := is.Type.(*ast.TypeSpec_Interface).Interface

		objFields.List = mergeFields(objFields.List, it.Fields.List)
	}
}

func mergeFields(objFields, intFields []*ast.Field) []*ast.Field {
	var found bool
	for _, ifield := range intFields {
		found = false

		for _, ofield := range objFields {
			if ifield.Name.Name == ofield.Name.Name {
				found = true
				break
			}
		}

		if found {
			continue
		}

		objFields = append(objFields, ifield)
	}

	return objFields
}
