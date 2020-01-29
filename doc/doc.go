// Package doc contains a CommonMark documentation generator for GraphQL Documents.
package doc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"gitlab.com/zaba505/markdown"
	"io"
	"path/filepath"
	"sync"
)

// Options contains the options for the Documentation generator.
type Options struct {
	Title string `json:"title"`
	HTML  bool   `json:"html"`

	toc *[]string
}

const (
	schema    = "schema"
	scalar    = "scalar"
	object    = "object"
	inter     = "interface"
	union     = "union"
	enum      = "enum"
	input     = "input"
	directive = "directive"
)

type declType uint16

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

func (o *Options) addContent(name string, typ, mask declType) declType {
	if mask&typ != 0 {
		switch typ {
		case schemaType:
		case scalarType:
			*o.toc = append(*o.toc, scalar)
		case objectType:
			*o.toc = append(*o.toc, object)
		case interType:
			*o.toc = append(*o.toc, inter)
		case unionType:
			*o.toc = append(*o.toc, union)
		case enumType:
			*o.toc = append(*o.toc, enum)
		case inputType:
			*o.toc = append(*o.toc, input)
		case directiveType:
			*o.toc = append(*o.toc, directive)
		}

		mask &= ^typ
	}

	*o.toc = append(*o.toc, name)

	return mask
}

// Generator generates CommonMark documentation for GraphQL Documents.
type Generator struct {
	sync.Mutex
	bytes.Buffer

	indent []byte

	mdOnce sync.Once
	md     *markdown.Markdown
}

// Reset overrides the bytes.Buffer Reset method to assist in cleaning up some Generator state.
func (g *Generator) Reset() {
	g.Buffer.Reset()
	if g.indent == nil {
		g.indent = make([]byte, 0, 2)
	}
	g.indent = g.indent[0:0]
}

// Generate generates CommonMark documentation for the given document.
func (g *Generator) Generate(ctx context.Context, doc *ast.Document, opts string) (err error) {
	g.Lock()
	defer func() {
		if err != nil {
			err = gen.GeneratorError{
				DocName: doc.Name,
				GenName: "doc",
				Msg:     err.Error(),
			}
		}
	}()
	defer g.Unlock()
	g.Reset()

	// Get generator options
	gOpts, oerr := getOptions(doc, opts)
	if oerr != nil {
		return oerr
	}

	// Generate types
	g.generateTypes(doc.Types, gOpts)

	// Extract generator context
	gCtx := gen.Context(ctx)

	// Open .md file
	base := doc.Name[:len(doc.Name)-len(filepath.Ext(doc.Name))]
	docFile, err := gCtx.Open(base + ".md")
	defer docFile.Close()
	if err != nil {
		return
	}

	// Write Title and Table of Contents
	_, err = writeToC(docFile, gOpts)
	if err != nil {
		return
	}

	// Write markdown
	b := g.Bytes()
	_, err = docFile.Write(b)
	if err != nil {
		return
	}

	if !gOpts.HTML {
		return
	}

	// Make sure markdown renderer is set
	if g.md == nil {
		g.mdOnce.Do(func() { g.md = markdown.New() })
	}

	// Open HTML file
	htmlFile, err := gCtx.Open(base + ".html")
	defer htmlFile.Close()
	if err != nil {
		return
	}

	err = g.md.Render(htmlFile, b)
	return
}

func noopGen(*ast.TypeSpec) {}

func (g *Generator) generateTypes(types []*ast.TypeDecl, opts *Options) {
	var fieldsBuf bytes.Buffer
	var typ declType
	var ts *ast.TypeSpec
	mask := schemaType | scalarType | objectType | interType | unionType | enumType | inputType | directiveType
	tLen := len(types) - 1
	for i, decl := range types {
		d, ok := decl.Spec.(*ast.TypeDecl_TypeSpec)
		if !ok {
			panic("only expected type spec and not type ext specs.")
		}
		ts = d.TypeSpec

		// Add to Table of Contents
		name := schema
		if ts.Name != nil {
			name = ts.Name.Name
		}

		// Generate type
		gen := noopGen
		switch v := ts.Type.(type) {
		case *ast.TypeSpec_Schema:
			if mask&schemaType != 0 {
				g.writeSectionHeader("Schema")
			}
			if v.Schema.RootOps == nil {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.P("*Root Operations*:")
				g.generateFields(v.Schema.RootOps.List, &fieldsBuf)
			}

			typ = schemaType
		case *ast.TypeSpec_Scalar:
			if mask&scalarType != 0 {
				g.writeSectionHeader("Scalar")
				g.WriteByte('\n')
			}

			typ = scalarType
		case *ast.TypeSpec_Object:
			if mask&objectType != 0 {
				g.writeSectionHeader("Object")
				g.WriteByte('\n')
			}

			gen = g.generateObject

			typ = objectType
		case *ast.TypeSpec_Interface:
			if mask&interType != 0 {
				g.writeSectionHeader("Interface")
				g.WriteByte('\n')
			}
			if v.Interface.Fields == nil {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.P("*Fields*:")
				g.generateFields(v.Interface.Fields.List, &fieldsBuf)
			}

			typ = interType
		case *ast.TypeSpec_Union:
			if mask&unionType != 0 {
				g.writeSectionHeader("Union")
				g.WriteByte('\n')
			}
			if len(v.Union.Members) == 0 {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.WriteString("*Members*: ")
				mLen := len(v.Union.Members) - 1
				for i, m := range v.Union.Members {
					g.WriteByte('*')
					g.WriteByte('*')
					g.WriteByte('[')
					g.WriteString(m.Name)
					g.WriteByte(']')
					g.WriteByte('(')
					g.WriteByte('#')
					g.WriteString(m.Name)
					g.WriteByte(')')
					g.WriteByte('*')
					g.WriteByte('*')
					if i != mLen {
						g.WriteByte(',')
						g.WriteByte(' ')
					}
				}
				g.WriteByte('\n')
			}

			typ = unionType
		case *ast.TypeSpec_Enum:
			if mask&enumType != 0 {
				g.writeSectionHeader("Enum")
				g.WriteByte('\n')
			}
			if v.Enum.Values == nil {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.P("*Values*:")
				g.generateFields(v.Enum.Values.List, &fieldsBuf)
			}

			typ = enumType
		case *ast.TypeSpec_Input:
			if mask&inputType != 0 {
				g.writeSectionHeader("Input")
				g.WriteByte('\n')
			}
			if v.Input.Fields == nil {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.P("*Fields*:")
				g.generateArgs(v.Input.Fields.List, &fieldsBuf)
			}

			typ = inputType
		case *ast.TypeSpec_Directive:
			if mask&directiveType != 0 {
				g.writeSectionHeader("Directive")
				g.WriteByte('\n')
			}
			if v.Directive.Args == nil {
				break
			}

			gen = func(_ *ast.TypeSpec) {
				g.WriteByte('\n')
				g.P("*Args*:")
				g.generateArgs(v.Directive.Args.List, &fieldsBuf)
			}

			typ = directiveType
		default:
			panic("unknown type spec type")
		}

		mask = opts.addContent(name, typ, mask)
		if typ != schemaType {
			g.writeTypeHeader(name)
		}

		if len(ts.Directives) > 0 {
			g.Write(g.indent)
			g.WriteString("*Directives*: ")
			g.writeDirectives(ts.Directives)
			g.WriteByte('\n')
		}

		decl.Doc.TextTo(&g.Buffer)

		gen(ts)

		if i != tLen {
			g.WriteByte('\n')
		}
	}
}

var (
	schemaName, schemaLink       = []byte("Schema"), []byte("](#Schema")
	scalarName, scalarLink       = []byte("Scalar"), []byte("s](#Scalar")
	objectName, objectLink       = []byte("Object"), []byte("s](#Object")
	interName, interLink         = []byte("Interface"), []byte("s](#Interface")
	unionName, unionLink         = []byte("Union"), []byte("s](#Union")
	enumName, enumLink           = []byte("Enum"), []byte("s](#Enum")
	inputName, inputLink         = []byte("Input"), []byte("s](#Input")
	directiveName, directiveLink = []byte("Directive"), []byte("s](#Directive")
)

// writeToC writes the Title and Table of Contents to the given io.Writer.
func writeToC(w io.Writer, opts *Options) (int64, error) {
	var b bytes.Buffer
	b.Grow(bytes.MinRead)

	// Title
	b.WriteByte('#')
	b.WriteByte(' ')
	b.WriteString(opts.Title)
	b.WriteByte('\n')

	// Generated line
	b.WriteString("*This was generated by gqlc.*")
	b.WriteByte('\n')
	b.WriteByte('\n')

	// Table of Contents
	b.WriteString("## Table of Contents")
	b.WriteByte('\n')

	name, link := make([]byte, 0, 20), make([]byte, 0, 23) // Assume longest would be a 5 character CJK unicode name => 4byte * 5char = 20
	listTok := []byte{'-', '*'}
	indent := make([]byte, 0, 2)
	for _, s := range *opts.toc {
		name, link = name[:0], link[:0]

		var shouldIndent, addS bool
		switch s {
		case schema:
			name = append(name, schemaName...)
			link = append(link, schemaLink...)
		case scalar:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}

			name = append(name, scalarName...)
			link = append(link, scalarLink...)
			shouldIndent = true
			addS = true
		case object:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}

			name = append(name, objectName...)
			link = append(link, objectLink...)
			shouldIndent = true
			addS = true
		case inter:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}

			name = append(name, interName...)
			link = append(link, interLink...)
			shouldIndent = true
			addS = true
		case union:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}
			name = append(name, unionName...)
			link = append(link, unionLink...)
			shouldIndent = true
			addS = true
		case enum:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}
			name = append(name, enumName...)
			link = append(link, enumLink...)
			shouldIndent = true
			addS = true
		case input:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}
			name = append(name, inputName...)
			link = append(link, inputLink...)
			shouldIndent = true
			addS = true
		case directive:
			if len(indent) > 0 {
				indent = indent[:len(indent)-1]
			}
			name = append(name, directiveName...)
			link = append(link, directiveLink...)
			shouldIndent = true
			addS = true
		default:
			name = append(name, s...)
			link = append(link, ']', '(', '#')
			link = append(link, name...)
		}

		b.Write(indent)
		b.WriteByte(listTok[len(indent)%2])
		b.WriteByte(' ')
		b.WriteByte('[')

		b.Write(name)

		b.Write(link)

		if addS {
			b.WriteByte('s')
		}

		b.WriteByte(')')
		b.WriteByte('\n')

		if shouldIndent {
			indent = append(indent, '\t')
		}
	}
	b.WriteByte('\n')

	return b.WriteTo(w)
}

func (g *Generator) writeSectionHeader(section string) {
	g.WriteByte('#')
	g.WriteByte('#')

	g.WriteByte(' ')
	g.WriteString(section)

	if section != "Schema" {
		g.WriteByte('s')
	}
	g.WriteByte('\n')
}

func (g *Generator) writeTypeHeader(name string) {
	g.WriteByte('#')
	g.WriteByte('#')
	g.WriteByte('#')

	g.WriteByte(' ')
	g.WriteString(name)

	g.WriteByte('\n')
}

func (g *Generator) writeDirectives(directives []*ast.DirectiveLit) {
	dLen := len(directives) - 1
	for i, d := range directives {
		g.WriteByte('@')
		g.WriteString(d.Name)

		if d.Args != nil {
			g.WriteByte('(')

			aLen := len(d.Args.Args) - 1
			for j, a := range d.Args.Args {
				g.WriteString(a.Name.Name)
				g.WriteByte(':')
				g.WriteByte(' ')

				var aVal interface{}
				switch v := a.Value.(type) {
				case *ast.Arg_BasicLit:
					aVal = v.BasicLit
				case *ast.Arg_CompositeLit:
					aVal = v.CompositeLit
				}
				g.printVal(aVal)

				if j != aLen {
					g.WriteByte(',')
					g.WriteByte(' ')
				}
			}

			g.WriteByte(')')
		}

		if i != dLen {
			g.WriteByte(',')
			g.WriteByte(' ')
		}
	}
	g.WriteByte('\n')
}

func (g *Generator) generateObject(ts *ast.TypeSpec) {
	obj := ts.Type.(*ast.TypeSpec_Object).Object

	if len(obj.Interfaces) > 0 {
		g.WriteByte('\n')
		g.Write(g.indent)
		g.WriteString("*Interfaces*: ")

		iLen := len(obj.Interfaces) - 1
		for i, inter := range obj.Interfaces {
			g.WriteString(inter.Name)

			if i != iLen {
				g.WriteByte(',')
				g.WriteByte(' ')
			}
		}
		g.WriteByte('\n')
	}

	if obj.Fields == nil {
		return
	}
	if len(obj.Fields.List) == 0 {
		return
	}

	var b bytes.Buffer
	g.WriteByte('\n')
	g.P("*Fields*:")
	g.generateFields(obj.Fields.List, &b)
}

// generateFields only generates a list of fields. It assumes any "Fields" section/list header
// has been generated.
//
func (g *Generator) generateFields(fields []*ast.Field, b *bytes.Buffer) {
	for _, f := range fields {
		b.Reset()

		// Write name
		g.Write(g.indent)
		g.WriteByte('-')
		g.WriteByte(' ')
		g.WriteString(f.Name.Name)

		// Write type
		if f.Type != nil {
			g.WriteByte(' ')
			g.WriteByte('*')
			g.WriteByte('*')
			g.WriteByte('(')
			var typ interface{}
			switch v := f.Type.(type) {
			case *ast.Field_Ident:
				typ = v.Ident
			case *ast.Field_List:
				typ = v.List
			case *ast.Field_NonNull:
				typ = v.NonNull
			}
			g.printType(typ)
			g.WriteByte(')')
			g.WriteByte('*')
			g.WriteByte('*')
		}
		g.WriteByte('\n')

		g.In()

		if len(f.Directives) > 0 {
			g.WriteByte('\n')
			g.Write(g.indent)
			g.WriteString("*Directives*: ")
			g.writeDirectives(f.Directives)
		}

		// Write descr
		f.Doc.TextTo(b)
		if b.Len() > 0 {
			g.WriteByte('\n')
			g.Write(g.indent)
			b.WriteTo(g)
		}

		// Write args
		if f.Args != nil {
			g.WriteByte('\n')
			g.P("*Args*:")
			g.generateArgs(f.Args.List, b)
		}

		g.Out()
	}
}

func (g *Generator) generateArgs(args []*ast.InputValue, b *bytes.Buffer) {
	for _, f := range args {
		b.Reset()

		// Write name
		g.Write(g.indent)
		g.WriteByte('-')
		g.WriteByte(' ')
		g.WriteString(f.Name.Name)

		// Write type
		if f.Type != nil {
			g.WriteByte(' ')
			g.WriteByte('*')
			g.WriteByte('*')
			g.WriteByte('(')
			var typ interface{}
			switch v := f.Type.(type) {
			case *ast.InputValue_Ident:
				typ = v.Ident
			case *ast.InputValue_List:
				typ = v.List
			case *ast.InputValue_NonNull:
				typ = v.NonNull
			}
			g.printType(typ)
			g.WriteByte(')')
			g.WriteByte('*')
			g.WriteByte('*')
		}
		g.WriteByte('\n')

		g.In()

		if len(f.Directives) > 0 {
			g.WriteByte('\n')
			g.Write(g.indent)
			g.WriteString("*Directives*: ")
			g.writeDirectives(f.Directives)
		}

		// Write descr
		f.Doc.TextTo(b)
		if b.Len() > 0 {
			g.WriteByte('\n')
			g.Write(g.indent)
			b.WriteTo(g)
		}

		// Write default value
		if f.Default != nil {
			g.WriteByte('\n')
			g.Write(g.indent)
			g.WriteString("*Default Value*")
			g.WriteByte(':')
			g.WriteByte(' ')

			g.WriteByte('`')
			var dv interface{}
			switch v := f.Default.(type) {
			case *ast.InputValue_BasicLit:
				dv = v.BasicLit
			case *ast.InputValue_CompositeLit:
				dv = v.CompositeLit
			}
			g.printVal(dv)
			g.WriteByte('`')
			g.WriteByte('\n')
		}

		g.Out()
	}
}

func (g *Generator) printType(typ interface{}) {
	switch v := typ.(type) {
	case *ast.Ident:
		switch v.Name {
		case "Int", "Float", "String", "Boolean", "ID":
			g.WriteString(v.Name)
		default:
			g.WriteByte('[')
			g.WriteString(v.Name)
			g.WriteByte(']')

			g.WriteByte('(')
			g.WriteByte('#')
			g.WriteString(v.Name)
			g.WriteByte(')')
		}
	case *ast.List:
		g.WriteByte('[')

		var nt interface{}
		switch w := v.Type.(type) {
		case *ast.List_Ident:
			nt = w.Ident
		case *ast.List_List:
			nt = w.List
		case *ast.List_NonNull:
			nt = w.NonNull
		}
		g.printType(nt)

		g.WriteByte(']')
	case *ast.NonNull:
		var nt interface{}
		switch w := v.Type.(type) {
		case *ast.NonNull_Ident:
			nt = w.Ident
		case *ast.NonNull_List:
			nt = w.List
		}
		g.printType(nt)

		g.WriteByte('!')
	}
}

// printVal prints a value
func (g *Generator) printVal(val interface{}) {
	switch v := val.(type) {
	case *ast.BasicLit:
		g.WriteString(v.Value)
	case *ast.ListLit:
		g.WriteByte('[')

		var vals []interface{}
		switch w := v.List.(type) {
		case *ast.ListLit_BasicList:
			for _, bval := range w.BasicList.Values {
				vals = append(vals, bval)
			}
		case *ast.ListLit_CompositeList:
			for _, cval := range w.CompositeList.Values {
				vals = append(vals, cval)
			}
		}

		vLen := len(vals) - 1
		for i, iv := range vals {
			g.printVal(iv)
			if i != vLen {
				g.WriteByte(',')
				g.WriteByte(' ')
			}
		}

		g.WriteByte(']')
	case *ast.ObjLit:
		g.WriteByte('{')
		g.WriteByte(' ')

		pLen := len(v.Fields) - 1
		for i, p := range v.Fields {
			g.WriteString(p.Key.Name)
			g.WriteString(": ")

			g.printVal(p.Val)

			if i != pLen {
				g.WriteByte(',')
			}
			g.WriteByte(' ')
		}

		g.WriteByte('}')
	case *ast.CompositeLit:
		switch w := v.Value.(type) {
		case *ast.CompositeLit_BasicLit:
			g.printVal(w.BasicLit)
		case *ast.CompositeLit_ListLit:
			g.printVal(w.ListLit)
		case *ast.CompositeLit_ObjLit:
			g.printVal(w.ObjLit)
		}
	}
}

// P prints the arguments to the generated output.
func (g *Generator) P(str ...interface{}) {
	g.Write(g.indent)
	for _, s := range str {
		switch v := s.(type) {
		case string:
			g.WriteString(v)
		case bool:
			fmt.Fprint(g, v)
		case int:
			fmt.Fprint(g, v)
		case float64:
			fmt.Fprint(g, v)
		}
	}
	g.WriteByte('\n')
}

// In increases the indent.
func (g *Generator) In() {
	g.indent = append(g.indent, '\t')
}

// Out decreases the indent.
func (g *Generator) Out() {
	if len(g.indent) > 0 {
		g.indent = g.indent[:len(g.indent)-1]
	}
}

// getOptions returns a generator options struct given all generator option metadata from the Doc and CLI.
// Precedence: CLI over Doc over Default
//
func getOptions(doc *ast.Document, opts string) (gOpts *Options, err error) {
	toc := make([]string, 0, len(doc.Types)+9)
	gOpts = &Options{
		Title: `Documentation`,
		toc:   &toc,
	}

	// Extract document directive options
	for _, d := range doc.Directives {
		if d.Name != "doc" {
			continue
		}

		if d.Args == nil {
			break
		}

		docOpts := d.Args.Args[0].Value.(*ast.Arg_CompositeLit).CompositeLit.Value.(*ast.CompositeLit_ObjLit).ObjLit
		for _, arg := range docOpts.Fields {
			switch arg.Key.Name {
			case "title":
				gOpts.Title = arg.Val.Value.(*ast.CompositeLit_BasicLit).BasicLit.Value
			case "html":
				v := arg.Val.Value.(*ast.CompositeLit_BasicLit).BasicLit.Value
				if v == "true" {
					gOpts.HTML = true
				}
			}
		}
	}

	// Unmarshal cli options
	if len(opts) > 0 {
		err = json.Unmarshal([]byte(opts), gOpts)
		if err != nil {
			return
		}
	}

	// Trim '"' from beginning and end of title string
	if gOpts.Title[0] == '"' {
		gOpts.Title = gOpts.Title[1 : len(gOpts.Title)-1]
	}
	return
}
