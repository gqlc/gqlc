package sort

import (
	"github.com/gqlc/graphql/ast"
)

// DeclType defines the order of types in the .md file
type DeclType uint16

const (
	SchemaType DeclType = 1 << iota
	ScalarType
	ObjectType
	InterType
	UnionType
	EnumType
	InputType
	DirectiveType
	ExtendType
)

type TypeSlice []*ast.TypeDecl

func (s TypeSlice) Len() int { return len(s) }
func (s TypeSlice) Less(i, j int) bool {
	it, jt := s[i], s[j]

	var its, jts *ast.TypeSpec
	itd, iok := it.Spec.(*ast.TypeDecl_TypeSpec)
	jtd, jok := jt.Spec.(*ast.TypeDecl_TypeSpec)

	if iok != jok { // TypeSpec < TypeExt
		return !iok && jok
	}

	if !iok && !jok {
		its, jts = it.Spec.(*ast.TypeDecl_TypeExtSpec).TypeExtSpec.Type, jt.Spec.(*ast.TypeDecl_TypeExtSpec).TypeExtSpec.Type
	} else {
		its, jts = itd.TypeSpec, jtd.TypeSpec
	}

	// Schema < Scalar < Object < Interface < Union < Enum < Input < Directive
	var iType, jType DeclType
	switch its.Type.(type) {
	case *ast.TypeSpec_Schema:
		iType = SchemaType
	case *ast.TypeSpec_Scalar:
		iType = ScalarType
	case *ast.TypeSpec_Object:
		iType = ObjectType
	case *ast.TypeSpec_Interface:
		iType = InterType
	case *ast.TypeSpec_Union:
		iType = UnionType
	case *ast.TypeSpec_Enum:
		iType = EnumType
	case *ast.TypeSpec_Input:
		iType = InputType
	case *ast.TypeSpec_Directive:
		iType = DirectiveType
	}
	switch jts.Type.(type) {
	case *ast.TypeSpec_Schema:
		jType = SchemaType
	case *ast.TypeSpec_Scalar:
		jType = ScalarType
	case *ast.TypeSpec_Object:
		jType = ObjectType
	case *ast.TypeSpec_Interface:
		jType = InterType
	case *ast.TypeSpec_Union:
		jType = UnionType
	case *ast.TypeSpec_Enum:
		jType = EnumType
	case *ast.TypeSpec_Input:
		jType = InputType
	case *ast.TypeSpec_Directive:
		jType = DirectiveType
	}

	if iType != jType {
		return iType < jType
	}

	return its.Name.Name < jts.Name.Name
}
func (s TypeSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
