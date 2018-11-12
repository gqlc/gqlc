// Package file declares the types used to represent a GraphQL IDL source.
//
package file

import (
	"gqlc/graphql/token"
	"strings"
)

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Decl interface {
	Node
	declNode()
}

// A Field represents a Field declaration in a GraphQL type declaration
// or an argument declaration in an arguments declaration.
// Field.Names is nil for unnamed parameters (parameter lists which only contain types)
// and embedded struct fields. In the latter case, the field name is the type name.
//
type Field struct {
	Doc     *DocGroup       // associated documentation; or nil
	Name    *Ident          // field/parameter names; or nil
	Args    CallExpr        // field arguments; or nil
	Type    Expr            // field/parameter type
	Default *BasicLit       // parameter default value; or nil
	Dirs    []*DirectiveLit // directives; or nil
}

func (f *Field) Pos() token.Pos {
	if f.Name != nil {
		return f.Name.Pos()
	}
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	if f.Dirs != nil {
		return f.Dirs[0].End()
	}
	return f.Type.End()
}

// A FieldList represents a list of Fields, enclosed by parentheses or braces.
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

// NumFields returns the number of parameters or struct fields represented by a FieldList.
func (f *FieldList) NumFields() (n int) {
	if f != nil {
		n = len(f.List)
	}
	return
}

// An expression is represented by a tree consisting of one
// or more of the following concrete expression nodes.
//
type (
	// A BadExpr node is a placeholder for expressions containing
	// syntax errors for which no correct expression nodes can be
	// created.
	BadExpr struct {
		From, To token.Pos
	}

	// An Ident node represents an identifier.
	Ident struct{}

	// A BasicList node represents a literal of basic type.
	BasicLit struct {
		ValuePos token.Pos   // literal position
		Kind     token.Token // token.INT, token.FLOAT, or token.STRING
		Value    string
	}

	DirectiveLit struct{}

	// A CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Lparen token.Pos // position of '('
		Args   []*Field  // arguments; or nil
		Rparen token.Pos // position of ')'
	}
)

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
//
type (
	SchemaType struct {
		Schema token.Pos // position of "schema" keyword
		Fields *FieldList
	}

	ScalarType struct {
		Scalar token.Pos // position of "scalar" keyword
	}

	ObjectType struct {
		Object token.Pos // position of "type" keyword
		Impls  []Expr    // implemented interfaces; or nil
		Fields *FieldList
	}

	InterfaceType struct {
		Interface token.Pos // position of "interface" keyword
		Fields    *FieldList
	}

	UnionType struct {
		Union   token.Pos // position of "union" keyword
		Members []Expr
	}

	EnumType struct {
		Enum   token.Pos // position of "enum" keyword
		Fields *FieldList
	}

	InputType struct {
		Input  token.Pos // position of "input" keyword
		Fields *FieldList
	}

	DirectiveType struct {
		Directive token.Pos // position of "directive" keyword
		Args      *CallExpr
		OnPos     token.Pos // position of "on" keyword
		Locs      []Expr    // valid locations where this directive can be applied
	}

	Extension struct {
		Extend token.Pos // position of "extend" keyword
		Type   *TypeSpec // the extended type
	}
)

// TODO: Handle extension situations i.e. no fields, no impls

func (x *BadExpr) Pos() token.Pos       { return x.From }
func (x *Ident) Pos() token.Pos         { return 0 } // TODO
func (x *BasicLit) Pos() token.Pos      { return 0 } // TODO
func (x *DirectiveLit) Pos() token.Pos  { return 0 } // TODO
func (x *SchemaType) Pos() token.Pos    { return x.Schema }
func (x *ScalarType) Pos() token.Pos    { return x.Scalar }
func (x *ObjectType) Pos() token.Pos    { return x.Object }
func (x *InterfaceType) Pos() token.Pos { return x.Interface }
func (x *UnionType) Pos() token.Pos     { return x.Union }
func (x *EnumType) Pos() token.Pos      { return x.Enum }
func (x *InputType) Pos() token.Pos     { return x.Input }
func (x *DirectiveType) Pos() token.Pos { return x.Directive }
func (x *Extension) Pos() token.Pos     { return x.Extend }

func (x *BadExpr) End() token.Pos       { return x.To }
func (x *Ident) End() token.Pos         { return 0 } // TODO
func (x *BasicLit) End() token.Pos      { return 0 } // TODO
func (x *DirectiveLit) End() token.Pos  { return 0 } // TODO
func (x *SchemaType) End() token.Pos    { return x.Fields.End() }
func (x *ScalarType) End() token.Pos    { return token.NoPos }
func (x *ObjectType) End() token.Pos    { return x.Fields.End() }
func (x *InterfaceType) End() token.Pos { return x.Fields.End() }
func (x *UnionType) End() token.Pos     { return x.Members[0].End() }
func (x *EnumType) End() token.Pos      { return x.Fields.End() }
func (x *InputType) End() token.Pos     { return x.Fields.End() }
func (x *DirectiveType) End() token.Pos { return x.Locs[0].End() }
func (x *Extension) End() token.Pos     { return x.Type.End() }

func (*BadExpr) exprNode()  {}
func (*Ident) exprNode()    {}
func (*BasicLit) exprNode() {}

type (
	// The Spec type stands for any of *ImportSpec, and *TypeSpec
	Spec interface {
		Node
		specNode()
	}

	// An ImportSpec node represents a single file import.
	ImportSpec struct {
		Doc    *DocGroup // associated documentation; or nil
		Name   *Ident    // local import name (including "."); or nil
		Path   *BasicLit // import path
		EndPos token.Pos // end of spec (overrides Path.Pos if nonzero)
	}

	// A TypeSpec node represents a GraphQL type declaration.
	TypeSpec struct {
		Doc  *DocGroup // associated documentation; or nil
		Name *Ident    // type name; or nil
		Dirs []Expr    // applied directives; or nil
		Type Expr      // *Ident, or any of the *XxxTypes
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
func (s *TypeSpec) End() token.Pos {
	e := s.Type.End()
	if e == token.NoPos {
		return s.Name.End()
	}
	return e
}

// specNode() ensures that only spec nodes can be
// assigned to a Spec.
//
func (*ImportSpec) specNode() {}
func (*TypeSpec) specNode()   {}

type (
	BadDecl struct {
		From, To token.Pos
	}

	GenDecl struct {
		Doc    *DocGroup   // associated documentation; or nil
		TokPos token.Pos   // position of Tok
		Tok    token.Token // IMPORT, TYPE_KEYWORD (e.g. schema, input, union)
		Lparen token.Pos   // position of '(', if any
		Specs  []Spec
		Rparen token.Pos // position of ')', if any
	}
)

func (d *BadDecl) Pos() token.Pos { return d.From }
func (d *GenDecl) Pos() token.Pos { return d.TokPos }

func (d *BadDecl) End() token.Pos { return d.To }
func (d *GenDecl) End() token.Pos {
	if d.Rparen.IsValid() {
		return d.Rparen + 1
	}
	return d.Specs[0].End()
}

func (*BadDecl) declNode() {}
func (*GenDecl) declNode() {}

// Doc represents a single line documentation source i.e. Description or Comment.
//
type Doc struct {
	// Text is the text after the first '#'.
	Text string

	// Pos is the position of the first '#' or '"'.
	Char token.Pos
}

// DocGroup represents a sequence of docs
// with no other tokens and no empty lines between.
//
type DocGroup struct {
	List []*Doc // len(List) > 0
}

func isWhitespace(ch byte) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func stripTrailingWhitespace(s string) string {
	i := len(s)
	for i > 0 && isWhitespace(s[i-1]) {
		i--
	}
	return s[0:i]
}

// Text returns the text of the comment.
// Documentation markers (#, ", """), the first space of a line comment, and
// leading and trailing empty lines are removed. Multiple empty lines are
// reduced to one, and trailing space on lines is trimmed. Unless the result
// is empty, it is newline-terminated.
//
func (x *DocGroup) Text() string {
	if x == nil {
		return ""
	}
	comments := make([]string, len(x.List))
	for i, c := range x.List {
		comments[i] = c.Text
	}

	lines := make([]string, 0, 10) // most comments are less than 10 lines
	for _, c := range comments {
		// Remove comment markers.
		// The parser has given us exactly the comment text.
		switch c[1] {
		case '#':
			//-style comment (no newline at the end)
			c = c[2:]
			// strip first space - required for Example tests
			if len(c) > 0 && c[0] == ' ' {
				c = c[1:]
			}
		case '"':
			// """-style description TODO
		}

		// Split on newlines.
		cl := strings.Split(c, "\n")

		// Walk lines, stripping trailing white space and adding to list.
		for _, l := range cl {
			lines = append(lines, stripTrailingWhitespace(l))
		}
	}

	// Remove leading blank lines; convert runs of
	// interior blank lines to a single blank line.
	n := 0
	for _, line := range lines {
		if line != "" || n > 0 && lines[n-1] != "" {
			lines[n] = line
			n++
		}
	}
	lines = lines[0:n]

	// Add final "" entry to get trailing newline from Join.
	if n > 0 && lines[n-1] != "" {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// Descriptor represents a single parsed GraphQL IDL file.
type Descriptor struct {
	Name string    // file name, relative to root of source tree
	Doc  *DocGroup // associated documentation

	// Names of files imported by this file.
	Imports []string

	// Indexes of the public imported files in the dependency list above.
	PublicImports []string
}
