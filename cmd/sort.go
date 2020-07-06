// sort.go sorts type decls by type and name

package cmd

import (
	"sort"

	"github.com/gqlc/graphql/ast"
)

// sortTypeDecls sorts type declarations both lexigraphically and by type.
func sortTypeDecls(decls []*ast.TypeDecl) []*ast.TypeDecl {
	dtypes := typeSlice(decls)
	sort.Sort(dtypes)

	return dtypes
}

// declType defines the order of types in the .md file
type declType uint16

// Top-Level type declarations in GraphQL IDL.
const (
	schemaType declType = 1 << iota
	scalarType
	objectType
	interType
	unionType
	enumType
	inputType
	directiveType
	extendType
)

// typeSlice represents a list of GraphQL type declarations
type typeSlice []*ast.TypeDecl

func (s typeSlice) Len() int { return len(s) }
func (s typeSlice) Less(i, j int) bool {
	it, jt := s[i], s[j]

	var its, jts *ast.TypeSpec
	itd, iok := it.Spec.(*ast.TypeDecl_TypeSpec)
	jtd, jok := jt.Spec.(*ast.TypeDecl_TypeSpec)

	if iok != jok { // TypeSpec < TypeExt
		return !iok && jok
	}

	its, jts = itd.TypeSpec, jtd.TypeSpec
	if !iok && !jok {
		its, jts = it.Spec.(*ast.TypeDecl_TypeExtSpec).TypeExtSpec.Type, jt.Spec.(*ast.TypeDecl_TypeExtSpec).TypeExtSpec.Type
	}

	// Schema < Scalar < Object < Interface < Union < Enum < Input < Directive
	iType, jType := getDeclType(its.Type), getDeclType(jts.Type)

	if iType != jType {
		return iType < jType
	}

	return its.Name.Name < jts.Name.Name
}

func getDeclType(ts interface{}) (iType declType) {
	switch ts.(type) {
	case *ast.TypeSpec_Schema:
		iType = schemaType
	case *ast.TypeSpec_Scalar:
		iType = scalarType
	case *ast.TypeSpec_Object:
		iType = objectType
	case *ast.TypeSpec_Interface:
		iType = interType
	case *ast.TypeSpec_Union:
		iType = unionType
	case *ast.TypeSpec_Enum:
		iType = enumType
	case *ast.TypeSpec_Input:
		iType = inputType
	case *ast.TypeSpec_Directive:
		iType = directiveType
	}
	return
}

func (s typeSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
