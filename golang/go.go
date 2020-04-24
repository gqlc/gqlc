// Package golang contains a Go generator for GraphQL Documents.
package golang

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/graphql/ast"
	"github.com/gqlc/graphql/token"
)

// Options contains the options for the Go generator.
type Options struct {
	// Package is the go package this will belong to. (default: main)
	Package string

	// Copy descriptions to Go
	Descriptions bool
}

// Generator generates Go code for a GraphQL schema.
type Generator struct {
	sync.Mutex
	bytes.Buffer

	indent []byte
}

// Reset overrides the bytes.Buffer Reset method to assist in cleaning up some Generator state.
func (g *Generator) Reset() {
	g.Buffer.Reset()
	if g.indent == nil {
		g.indent = make([]byte, 0, 5)
	}
	g.indent = g.indent[0:0]
}

var typeSuffix = []byte("Type")

// Generate generates Go code for the given document.
func (g *Generator) Generate(ctx context.Context, doc *ast.Document, opts map[string]interface{}) (err error) {
	g.Lock()
	defer func() {
		if err != nil {
			err = gen.GeneratorError{
				DocName: doc.Name,
				GenName: "go",
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

	// Generate package and imports
	g.writeHeader(g, []byte(gOpts.Package))

	// Generate types
	totalTypes := len(doc.Types) - 1
	for i, d := range doc.Types {
		ts, ok := d.Spec.(*ast.TypeDecl_TypeSpec)
		if !ok {
			continue
		}
		if _, ok = ts.TypeSpec.Type.(*ast.TypeSpec_Schema); ok {
			continue
		}

		// Generate variable declaration
		name := ts.TypeSpec.Name.Name
		g.WriteString("var")
		g.WriteByte(' ')
		g.WriteString(name)
		g.Write(typeSuffix)
		g.WriteByte(' ')
		g.WriteByte('=')
		g.WriteByte(' ')
		g.WriteString("graphql")
		g.WriteByte('.')

		// Generate GraphQL*Type construction
		switch ts.TypeSpec.Type.(type) {
		case *ast.TypeSpec_Scalar:
			g.generateScalar(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Object:
			g.generateObject(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Interface:
			g.generateInterface(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Union:
			g.generateUnion(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Enum:
			g.generateEnum(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Input:
			g.generateInput(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		case *ast.TypeSpec_Directive:
			g.generateDirective(name, gOpts.Descriptions, d.Doc, ts.TypeSpec)
		}

		if i != totalTypes {
			g.P()
		}
	}

	// Extract generator context
	gCtx := gen.Context(ctx)

	// Open file to write to
	goFileName := doc.Name[:len(doc.Name)-len(filepath.Ext(doc.Name))]
	goFile, err := gCtx.Open(goFileName + ".go")
	defer goFile.Close()
	if err != nil {
		return
	}

	// Write generated output
	_, err = g.WriteTo(goFile)
	return
}

var (
	packagePrefix = []byte("package ")
	importStmt    = []byte(`import "github.com/graphql-go/graphql"`)
	newLines      = []byte{'\n', '\n'}
)

func (g *Generator) writeHeader(w io.Writer, packageName []byte) {
	w.Write(packagePrefix)
	w.Write(packageName)
	w.Write(newLines)

	w.Write(importStmt)
	w.Write(newLines)
}

func (g *Generator) generateScalar(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	g.P("NewScalar(graphql.ScalarConfig{")
	g.In()
	g.P("Name: \"", name, "\",")

	if doc != nil && descr {
		text := doc.Text()
		if len(text) > 0 {
			g.P("Description: \"", text[:len(text)-1], "\",")
		}
	}

	resolver := getResolver(ts.Directives)
	if resolver == "" {
		resolver = "func(value interface{}) interface{} { return nil }"
	}

	g.P("Serialize: ", resolver, ",")
	g.Out()

	g.P("})")
}

func (g *Generator) generateObject(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	obj := ts.Type.(*ast.TypeSpec_Object).Object

	g.P("NewObject(graphql.ObjectConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	// Print interfaces
	interLen := len(obj.Interfaces)
	if interLen == 1 {
		g.Write(g.indent)
		g.WriteString("Interfaces: []*graphql.Interface{ ")
		g.WriteString(obj.Interfaces[0].Name)
		g.Write(typeSuffix)
		g.WriteByte(' ')
		g.WriteByte('}')
		g.WriteByte(',')
		g.WriteByte('\n')
	}
	if interLen > 1 {
		g.Write(g.indent)
		g.WriteString("Interfaces: []*graphql.Interface{")
		g.WriteByte('\n')
		g.In()

		for _, inter := range obj.Interfaces {
			g.P(inter.Name, typeSuffix, ",")
		}

		g.Out()
		g.P("},")
	}

	g.P("Fields: graphql.Fields{")
	g.In()

	g.generateFields(obj.Fields, descr, true)

	g.Out()

	g.P("},")

	if doc != nil && descr {
		g.printDescr(doc)
		g.WriteByte('\n')
	}

	g.Out()
	g.P("})")
}

func (g *Generator) generateInterface(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	inter := ts.Type.(*ast.TypeSpec_Interface).Interface

	g.P("NewInterface(graphql.InterfaceConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	g.P("Fields: graphql.Fields{")
	g.In()

	g.generateFields(inter.Fields, descr, false)

	g.Out()

	g.P("},")

	if doc != nil && descr {
		g.printDescr(doc)
		g.WriteByte('\n')
	}

	g.Out()
	g.P("})")
}

func (g *Generator) generateFields(fields *ast.FieldList, descr, resolve bool) {
	for _, f := range fields.List {
		g.P('"', f.Name.Name, '"', ": &graphql.Field{")
		g.In()

		g.Write(g.indent)
		g.WriteString("Type: ")

		var fieldType interface{}
		switch v := f.Type.(type) {
		case *ast.Field_Ident:
			fieldType = v.Ident
		case *ast.Field_List:
			fieldType = v.List
		case *ast.Field_NonNull:
			fieldType = v.NonNull
		}
		g.printType(fieldType)
		g.WriteByte(',')
		g.WriteByte('\n')

		if f.Args != nil {
			g.P("Args: graphql.FieldConfigArgument{")
			g.In()

			for _, a := range f.Args.List {
				g.P('"', a.Name.Name, '"', ": &graphql.ArgumentConfig{")
				g.In()
				g.Write(g.indent)
				g.WriteString("Type: ")

				var argType interface{}
				switch v := a.Type.(type) {
				case *ast.InputValue_Ident:
					argType = v.Ident
				case *ast.InputValue_List:
					argType = v.List
				case *ast.InputValue_NonNull:
					argType = v.NonNull
				}
				g.printType(argType)
				g.WriteByte(',')
				g.WriteByte('\n')

				if a.Default != nil {
					g.Write(g.indent)
					g.WriteString("DefaultValue: ")

					var defType interface{}
					switch v := a.Default.(type) {
					case *ast.InputValue_BasicLit:
						defType = v.BasicLit
					case *ast.InputValue_CompositeLit:
						defType = v.CompositeLit
					}
					g.printVal(defType)
					g.WriteByte(',')
					g.WriteByte('\n')
				}

				if a.Doc != nil && descr {
					g.printDescr(a.Doc)
					g.WriteByte('\n')
				}

				g.Out()
				g.P("},")
			}

			g.Out()
			g.P("},")
		}

		if resolve {
			resolver := getResolver(f.Directives)
			if resolver == "" {
				resolver = "func(p graphql.ResolveParams) (interface{}, error) { return nil, nil }"
			}

			g.P("Resolve: ", resolver, ",")
		}

		if f.Doc != nil && descr {
			g.printDescr(f.Doc)
			g.WriteByte('\n')
		}

		g.Out()
		g.P("},")
	}
}

func (g *Generator) generateUnion(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	union := ts.Type.(*ast.TypeSpec_Union).Union

	g.P("NewUnion(graphql.UnionConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	// Print members
	memsLen := len(union.Members)
	if memsLen == 1 {
		g.P("Types: []*graphql.Object{ ", union.Members[0], typeSuffix, " },")
	}
	if memsLen > 1 {
		g.P("Types: []*graphql.Object{")
		g.In()

		for _, mem := range union.Members {
			g.P(mem.Name, typeSuffix, ',')
		}

		g.Out()
		g.P("},")
	}

	resolver := getResolver(ts.Directives)
	if resolver == "" {
		resolver = "func(p graphql.ResolveTypeParams) *graphql.Object { return nil }"
	}

	g.P("ResolveType: ", resolver, ",")

	if doc != nil && descr {
		g.printDescr(doc)
		g.WriteByte('\n')
	}

	g.Out()
	g.P("})")
}

func (g *Generator) generateEnum(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	enum := ts.Type.(*ast.TypeSpec_Enum).Enum

	g.P("NewEnum(graphql.EnumConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	if doc != nil && descr {
		g.printDescr(doc)
		g.WriteByte('\n')
	}

	g.P("Values: graphql.EnumValueConfigMap{")
	g.In()

	for _, v := range enum.Values.List {
		g.P('"', v.Name.Name, '"', ": &graphql.EnumValueConfig{")
		g.In()

		val := getValue(v.Directives)
		if val == "" {
			val = v.Name.Name
		}
		g.P("Value: \"", val, "\",")

		if v.Doc != nil && descr {
			g.printDescr(v.Doc)
			g.WriteByte('\n')
		}

		g.Out()

		g.P("},")
	}

	g.Out()
	g.P("},")

	g.Out()
	g.P("})")
}

func (g *Generator) generateInput(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	input := ts.Type.(*ast.TypeSpec_Input).Input

	g.P("NewInputObject(graphql.InputObjectConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	g.P("Fields: graphql.InputObjectConfigFieldMap{")
	g.In()

	for _, f := range input.Fields.List {
		g.P('"', f.Name.Name, '"', ": &graphql.InputObjectFieldConfig{")
		g.In()
		g.Write(g.indent)
		g.WriteString("Type: ")

		var fieldType interface{}
		switch v := f.Type.(type) {
		case *ast.InputValue_Ident:
			fieldType = v.Ident
		case *ast.InputValue_List:
			fieldType = v.List
		case *ast.InputValue_NonNull:
			fieldType = v.NonNull
		}
		g.printType(fieldType)
		g.WriteByte(',')
		g.WriteByte('\n')

		if f.Default != nil {
			g.Write(g.indent)
			g.WriteString("DefaultValue: ")

			var defType interface{}
			switch v := f.Default.(type) {
			case *ast.InputValue_BasicLit:
				defType = v.BasicLit
			case *ast.InputValue_CompositeLit:
				defType = v.CompositeLit
			}
			g.printVal(defType)
			g.WriteByte(',')
			g.WriteByte('\n')
		}

		g.Out()

		if f.Doc != nil && descr {
			g.printDescr(f.Doc)
			g.WriteByte('\n')
		}

		g.P("},")
	}

	g.Out()
	g.P("},")

	if doc != nil && descr {
		g.printDescr(doc)
		g.WriteByte('\n')
	}

	g.Out()
	g.P("})")
}

func (g *Generator) generateDirective(name string, descr bool, doc *ast.DocGroup, ts *ast.TypeSpec) {
	directive := ts.Type.(*ast.TypeSpec_Directive).Directive

	g.P("NewDirective(graphql.DirectiveConfig{")
	g.In()

	g.P("Name: \"", name, "\",")

	if doc != nil && descr {
		text := doc.Text()
		if len(text) > 0 {
			g.P("Description: \"", text[:len(text)-1], "\",")
		}
	}

	// Print locations
	locsLen := len(directive.Locs)
	if locsLen == 1 {
		g.P("Locations: []string{ ", directive.Locs[0].Loc.String(), " }")
	}
	if locsLen > 1 {
		g.P("Locations: []string{")
		g.In()

		for _, loc := range directive.Locs {
			g.P('"', loc.Loc.String(), "\",")
		}

		g.Out()
		g.P("},")
	}

	if directive.Args != nil {
		g.P("Args: graphql.FieldConfigArgument{")
		g.In()

		for _, a := range directive.Args.List {
			g.P('"', a.Name.Name, '"', ": &graphql.ArgumentConfig{")
			g.In()

			g.Write(g.indent)
			g.WriteString("Type: ")

			var fieldType interface{}
			switch v := a.Type.(type) {
			case *ast.InputValue_Ident:
				fieldType = v.Ident
			case *ast.InputValue_List:
				fieldType = v.List
			case *ast.InputValue_NonNull:
				fieldType = v.NonNull
			}
			g.printType(fieldType)
			g.WriteByte(',')
			g.WriteByte('\n')

			if a.Default != nil {
				g.Write(g.indent)
				g.WriteString("DefaultValue: ")

				var defType interface{}
				switch v := a.Default.(type) {
				case *ast.InputValue_BasicLit:
					defType = v.BasicLit
				case *ast.InputValue_CompositeLit:
					defType = v.CompositeLit
				}
				g.printVal(defType)
				g.WriteByte(',')
				g.WriteByte('\n')
			}

			if a.Doc != nil && descr {
				g.printDescr(a.Doc)
				g.WriteByte('\n')
			}

			g.Out()
			g.P("},")
		}

		g.Out()

		g.P("},")
	}

	g.Out()
	g.P("})")
}

func (g *Generator) printDescr(doc *ast.DocGroup) {
	text := doc.Text()
	if len(text) > 0 {
		g.Write(g.indent)
		g.WriteString("Description: \"")

		g.WriteString(text[:len(text)-1])

		g.WriteByte('"')
		g.WriteByte(',')
	}
}

// printType prints a field type
func (g *Generator) printType(typ interface{}) {
	switch v := typ.(type) {
	case *ast.Ident:
		name := v.Name

		switch name {
		case "Int":
			name = "graphql.Int"
		case "Float":
			name = "graphql.Float"
		case "String":
			name = "graphql.String"
		case "Boolean":
			name = "graphql.Boolean"
		case "ID":
			name = "graphql.ID"
		default:
			name = name + "Type"
		}

		g.WriteString(name)
	case *ast.List:
		g.WriteString("graphql.NewList(")

		switch w := v.Type.(type) {
		case *ast.List_Ident:
			typ = w.Ident
		case *ast.List_List:
			typ = w.List
		case *ast.List_NonNull:
			typ = w.NonNull
		}
		g.printType(typ)

		g.WriteByte(')')
	case *ast.NonNull:
		g.WriteString("graphql.NewNonNull(")

		switch w := v.Type.(type) {
		case *ast.NonNull_Ident:
			typ = w.Ident
		case *ast.NonNull_List:
			typ = w.List
		}
		g.printType(typ)

		g.WriteByte(')')
	}
}

// printVal prints a value
func (g *Generator) printVal(val interface{}) {
	switch v := val.(type) {
	case *ast.BasicLit:
		isEnum := v.Kind == token.Token_IDENT
		if isEnum {
			g.WriteRune('"')
		}
		g.WriteString(v.Value)
		if isEnum {
			g.WriteRune('"')
		}
	case *ast.ListLit:
		g.WriteString("[]interface{}{")

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

		g.WriteByte('}')
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
		case []byte:
			g.Write(v)
		case byte:
			g.WriteByte(v)
		case rune:
			g.WriteRune(v)
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
func getOptions(doc *ast.Document, opts map[string]interface{}) (gOpts *Options, err error) {
	gOpts = &Options{
		Package: "main",
	}

	// Extract document directive options
	for _, d := range doc.Directives {
		if d.Name != "go" {
			continue
		}

		if d.Args == nil {
			break
		}

		docOpts := d.Args.Args[0].Value.(*ast.Arg_CompositeLit).CompositeLit.Value.(*ast.CompositeLit_ObjLit).ObjLit
		for _, arg := range docOpts.Fields {
			switch arg.Key.Name {
			case "package":
				gOpts.Package = arg.Val.Value.(*ast.CompositeLit_BasicLit).BasicLit.Value
			case "descriptions":
				b, err := strconv.ParseBool(arg.Val.Value.(*ast.CompositeLit_BasicLit).BasicLit.Value)
				if err != nil {
					return gOpts, err
				}

				gOpts.Descriptions = b
			}
		}
	}

	// Unmarshal cli options
	if opts == nil {
		return
	}
	if p, ok := opts["package"]; ok {
		gOpts.Package, _ = p.(string)
	}
	if d, ok := opts["descriptions"]; ok {
		gOpts.Descriptions, _ = d.(bool)
	}

	// Trim '"' from beginning and end of title string
	if gOpts.Package[0] == '"' {
		gOpts.Package = gOpts.Package[1 : len(gOpts.Package)-1]
	}
	return
}

func getResolver(dirs []*ast.DirectiveLit) string {
	for _, d := range dirs {
		if d.Name != "resolver" {
			continue
		}

		return strings.Trim(d.Args.Args[0].Value.(*ast.Arg_BasicLit).BasicLit.Value, "\"")
	}
	return ""
}

func getValue(dirs []*ast.DirectiveLit) string {
	for _, d := range dirs {
		if d.Name != "as" {
			continue
		}

		return strings.Trim(d.Args.Args[0].Value.(*ast.Arg_BasicLit).BasicLit.Value, "\"")
	}
	return ""
}
