package ast

import "gqlc/sl/token"

// All node types implement the Node interface.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
	End() token.Pos // position of first character immediately after the node
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

// A Description node represents a single '"' of '"""' description
type Description struct {
	Quote token.Pos // position of '"' starting description
	Text  string    // description text
}

func (d *Description) Pos() token.Pos { return d.Quote }
func (d *Description) End() token.Pos { return token.Pos(int(d.Quote) + len(d.Text)) }

type Field struct{}

func (f *Field) Pos() token.Pos { return 0 }
func (f *Field) End() token.Pos { return 0 }

type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace, if any
}

func (f *FieldList) Pos() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Closing.IsValid() {
		return f.Closing + 1
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {

		}
	}
	return n
}

type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	//
	BadExpr struct {
		From, To token.Pos // position range of bad expression
	}

	// an Ident node represents an identifier.
	Ident struct {
		NamePos token.Pos // identifier position
		Name    string    // identifier name
	}

	// A BasicLit node represents a literal of basic type.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
		Value    string      // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// A CompositeLit node represents a composite literal.
	CompositeLit struct {
		Type       Expr      // literal type; or nil
		Lbrace     token.Pos // position of "{"
		Elts       []Expr    // list of composite elements; or nil
		Rbrace     token.Pos // position of "}"
		Incomplete bool      // true if (source) expressions are missing in the Elts list
	}

	// A ParenExpr node represents a parenthesized expression.
	ParenExpr struct {
		Lparen token.Pos // position of "("
		X      Expr      // parenthesized expression
		Rparen token.Pos // position of ")"
	}
)

type (
	// A ScalarType node represents a scalar type.
	ScalarType struct {
		Scalar     token.Pos // position of "scalar" keyword
		Directives []*FieldList
	}

	// A ObjectType node represents a object type.
	ObjectType struct {
		Object token.Pos // position of "type" keyword
	}

	// A InterfaceType node represents a interface type.
	InterfaceType struct {
		Interface token.Pos // position of "interface" keyword
	}

	// A UnionType node represents a union type.
	UnionType struct {
		Union token.Pos // position of "union" keyword
	}

	// A EnumType node represents a enum type.
	EnumType struct {
		Enum token.Pos // position of "enum" keyword
	}

	// A InputType node represents a input type.
	InputType struct {
		Input token.Pos // position of "input" keyword
	}

	// A DirectiveType node represents a directive type.
	DirectiveType struct {
		Directive token.Pos // position of "directive" keyword
	}

	// A ExtendType node represents a extend type.
	ExtendType struct {
		Extend token.Pos // position of "extend" keyword
	}
)

// Pos and End implementations for expression/type nodes.

func (x *BadExpr) Pos() token.Pos  { return x.From }
func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *CompositeLit) Pos() token.Pos {
	if x.Type != nil {
		return x.Type.Pos()
	}
	return x.Lbrace
}
func (x *ParenExpr) Pos() token.Pos { return x.Lparen }

func (x *ScalarType) Pos() token.Pos    { return x.Scalar }
func (x *ObjectType) Pos() token.Pos    { return x.Object }
func (x *InterfaceType) Pos() token.Pos { return x.Interface }
func (x *UnionType) Pos() token.Pos     { return x.Union }
func (x *EnumType) Pos() token.Pos      { return x.Enum }
func (x *InputType) Pos() token.Pos     { return x.Input }
func (x *DirectiveType) Pos() token.Pos { return x.Directive }
func (x *ExtendType) Pos() token.Pos    { return x.Extend }

// TODO
func (x *BadExpr) End() token.Pos      { return x.To }
func (x *Ident) End() token.Pos        { return 0 }
func (x *BasicLit) End() token.Pos     { return 0 }
func (x *CompositeLit) End() token.Pos { return 0 }
func (x *ParenExpr) End() token.Pos    { return 0 }

func (x *ScalarType) End() token.Pos    { return 0 }
func (x *ObjectType) End() token.Pos    { return 0 }
func (x *InterfaceType) End() token.Pos { return 0 }
func (x *UnionType) End() token.Pos     { return 0 }
func (x *EnumType) End() token.Pos      { return 0 }
func (x *InputType) End() token.Pos     { return 0 }
func (x *DirectiveType) End() token.Pos { return 0 }
func (x *ExtendType) End() token.Pos    { return 0 }

// exprNode() ensures that only expression/type nodes can be
// assigned to an Expr.
//
func (*BadExpr) exprNode()      {}
func (*Ident) exprNode()        {}
func (*BasicLit) exprNode()     {}
func (*CompositeLit) exprNode() {}
func (*ParenExpr) exprNode()    {}

func (*ScalarType) exprNode()    {}
func (*ObjectType) exprNode()    {}
func (*InterfaceType) exprNode() {}
func (*UnionType) exprNode()     {}
func (*EnumType) exprNode()      {}
func (*InputType) exprNode()     {}
func (*DirectiveType) exprNode() {}
func (*ExtendType) exprNode()    {}

type (
	// The Spec type stands for any of *ImportSpec or *TypeSpec
	Spec interface {
		Node
		specNode()
	}

	// An ImportSpec node represents a single package import.
	ImportSpec struct {
		Doc    *Description // associated documentation
		Name   *Ident       // local package name (including "."); or nil
		Path   *BasicLit    // import path
		EndPos token.Pos    // end of spec (overrides Path.Pos if nonzero)
	}

	TypeSpec struct {
		Doc  *Description // associated documentation
		Name *Ident       // type name
		Type Expr         // *Ident,
	}
)

// Pos and End implementations for spec nodes.

func (s *ImportSpec) Pos() token.Pos {
	if s.Name != nil {
		return s.Name.Pos()
	}
	return s.Path.Pos()
}
func (s *TypeSpec) Pos() token.Pos { return s.Name.Pos() }

func (s *ImportSpec) End() token.Pos {
	if s.EndPos != 0 {
		return s.EndPos
	}
	return s.Path.End()
}
func (s *TypeSpec) End() token.Pos { return s.Type.End() }

// specNode ensures that only spec nodes can be
// assigned to a spec.
//
func (s *ImportSpec) specNode() {}
func (s *TypeSpec) specNode()   {}

// A File node represents a GraphQL IDL source file.
//
type File struct{}
