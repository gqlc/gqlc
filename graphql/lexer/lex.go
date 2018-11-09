// Package lexer implements a lexer for GraphQL IDL source text.
//
package lexer

import (
	"fmt"
	"gqlc/graphql/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Item represents a lexed token.
type Item struct {
	Typ  token.Token
	Pos  token.Pos
	Val  string
	Line int
}

func (i Item) String() string {
	switch {
	case i.Typ == token.EOF:
		return "EOF"
	case i.Typ == token.ERR:
		return i.Val
	case i.Typ >= token.PACKAGE:
		return fmt.Sprintf("<%s>", i.Val)
	case len(i.Val) > 10:
		return fmt.Sprintf("%.10q...", i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}

// Interface defines the simplest API any consumer of a lexer could need.
type Interface interface {
	// NextItem returns the next lexed Item
	NextItem() Item

	// Drain drains the remaining items. Used only by parser if error occurs.
	Drain()
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lxr) stateFn

// A Mode value is a set of flags (or 0). They control lexer behaviour
type Mode uint

const (
	ScanComments Mode = 1 << iota // Return comments as COMMENT tokens
)

type lxr struct {
	// immutable state
	file *token.File
	name string
	mode Mode
	src  []byte

	// scanning state
	pos   int
	start int
	width int
	line  int
	items chan Item
}

// Lex lexs the given src based on the the GraphQL IDL specification.
func Lex(file *token.File, src []byte, mode Mode) Interface {
	l := &lxr{
		mode:  mode,
		file:  file,
		name:  file.Name(),
		src:   src,
		items: make(chan Item),
		line:  1,
	}

	go l.run()
	return l
}

const bom = 0xFEFF

// run runs the state machine for the lexer.
func (l *lxr) run() {
	r := l.next()
	if r == bom {
		l.ignore()
	} else {
		l.backup()
	}

	for state := lexDoc; state != nil; {
		state = state(l)
	}
	close(l.items)
}

const eof = -1

// next returns the next rune in the src.
func (l *lxr) next() rune {
	if int(l.pos) >= len(l.src) {
		l.width = 0
		return eof
	}

	r, w := utf8.DecodeRune(l.src[l.pos:])
	l.width = w
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the src.
func (l *lxr) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lxr) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.src[l.pos] == '\n' {
		l.line--
	}
}

// TODO: Check emitted value for newline characters and subtract them from l.line
// emit passes an item back to the client.
func (l *lxr) emit(t token.Token) {
	l.items <- Item{t, l.file.Pos(l.start), string(l.src[l.start:l.pos]), l.line}
	l.start = l.pos
}

// ignore skips over the pending src before this point.
func (l *lxr) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lxr) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lxr) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lxr) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{token.ERR, l.file.Pos(l.start), fmt.Sprintf(format, args...), l.line}
	return nil
}

// ignoreSpace consume all whitespace
func (l *lxr) ignoreSpace() {
	l.acceptRun(spaceChars)
	l.ignore()
}

// NextItem returns the next item from the src.
// Called by the parser, not in the lexing goroutine.
func (l *lxr) NextItem() Item {
	return <-l.items
}

// Drain drains the output so the lexing goroutine will exit.
// Called by the parser, not in the lexing goroutine.
func (l *lxr) Drain() {
	for range l.items {
	}
}

const spaceChars = " \t\r\n"

// ignoreComment is a helper for discarding document comments
func (l *lxr) ignoreComment() {
	for r := l.next(); !isEndOfLine(r) && r != eof; {
		r = l.next()
	}

	l.ignore()
}

// lexDoc scans a GraphQL schema language document
func lexDoc(l *lxr) stateFn {
	switch r := l.next(); {
	case r == eof:
		if l.pos > l.start {
			return l.errorf("unexpected eof")
		}
		l.emit(token.EOF)
		return nil
	case isSpace(r):
		l.ignoreSpace()
	case r == '#':
		if l.mode != ScanComments {
			l.ignoreComment()
			break
		}
		for s := l.next(); s != '\r' && s != '\n' && s != eof; {
			s = l.next()
		}
		l.emit(token.COMMENT)
	case r == '"':
		l.backup()
		if !l.scanString() {
			return l.errorf("bad string syntax: %q", l.src[l.start:l.pos])
		}
		l.emit(token.STRING)
	case isAlphaNumeric(r) && !unicode.IsDigit(r):
		l.backup()
		return lexImportsOrDef
	}
	return lexDoc
}

// lexImportsOrDef
func lexImportsOrDef(l *lxr) stateFn {
	// First, lex the identifier
	ident := l.scanIdentifier()
	if !ident.IsKeyword() {
		return l.errorf("invalid type declaration")
	}

	if ident != token.ON {
		l.emit(ident)
	}

	if ident != token.IMPORT && ident != token.EXTEND && ident != token.SCHEMA && ident != token.DIRECTIVE {
		l.acceptRun(" \t")
		l.ignore()

		if !isAlphaNumeric(l.peek()) {
			return l.errorf("expected type name for type decl: %s", ident)
		}

		name := l.scanIdentifier()
		l.emit(name)
	}

	switch ident {
	case token.IMPORT:
		return lexImports
	case token.SCALAR:
		return lexScalar
	case token.SCHEMA, token.TYPE, token.INTERFACE, token.ENUM, token.INPUT:
		return lexObject
	case token.UNION:
		return lexUnion
	case token.DIRECTIVE:
		return lexDirective
	case token.EXTEND:
		l.acceptRun(" \t")
		l.ignore()
		return lexImportsOrDef
	}

	return l.errorf("unknown type definition: %v", ident)
}

// lexImports scans an import block
func lexImports(l *lxr) stateFn {
	// import keyword has already been emitted. expect next is '('
	l.acceptRun(" \t")
	l.ignore()

	r := l.peek()
	if r != '(' && r != '"' {
		return l.errorf("missing ( or \" to begin import statement")
	}

	switch r {
	case '"':
		if !l.scanString() {
			return l.errorf("malformed src string: %s", l.src[l.start:l.pos])
		}
		l.emit(token.STRING)
	case '(':
		l.accept("(")
		l.emit(token.LPAREN)
		ok := l.scanList(")", defListSep, 0, func(ll *lxr) bool {
			if ll.scanString() {
				ll.emit(token.STRING)
				return true
			}
			ll.errorf("malformed import string")
			return false
		})
		if !ok {
			return nil
		}
		l.emit(token.RPAREN)
	}

	return lexDoc
}

// lexScalar scans a scalar type definition
func lexScalar(l *lxr) stateFn {
	l.acceptRun(" \t")
	l.ignore()

	r := l.peek()
	switch r {
	case '\n', '\r':
		break
	case '@':
		ok := l.scanDirectives("\r\n", " \t")
		if !ok {
			return nil
		}
	}
	return lexDoc
}

// lexObject scans a schema, object, interface, enum, or src type definition
func lexObject(l *lxr) stateFn {
	l.acceptRun(" \t")
	l.ignore()

	r := l.peek()
	switch r {
	case 'i':
		implIdent := l.scanIdentifier()
		if implIdent != token.IMPLEMENTS {
			return l.errorf("invalid identifier in object type signature")
		}
		l.emit(implIdent)

		ok := l.scanList("@{\n", "&", 0, func(ll *lxr) bool {
			name := ll.scanIdentifier()
			if name == token.ERR {
				ll.errorf("error occurred when scanning implements list")
				return false
			}
			ll.emit(name)
			return true
		})
		if !ok {
			return nil
		}
		l.backup()

		r = l.peek()
		switch r {
		case '@':
			goto dirs
		case '{':
			return lexFields
		}
		break
	dirs:
		fallthrough
	case '@':
		ok := l.scanDirectives("{\r\n", " \t")
		if !ok {
			return nil
		}
		l.backup()

		r = l.peek()
		if r == '\r' || r == '\n' || r == eof {
			break
		}

		fallthrough
	case '{':
		return lexFields
	default:
		return l.errorf("unexpected character encountered in object declaration: %s", string(l.src[l.pos]))
	}
	return lexDoc
}

// lexFieldDefs scans a FieldsDefinition, EnumValuesDefinition, InputFieldsDefinition list
func lexFields(l *lxr) stateFn {
	l.accept("{")
	l.emit(token.LBRACE)

	ok := l.scanList("}", defListSep, 0, func(ll *lxr) bool {
		if ll.accept("\"") {
			ll.backup()
			ok := ll.scanString()
			if !ok {
				return false
			}
			ll.emit(token.DESCRIPTION)
			ll.ignoreSpace()
		}

		tok := ll.scanIdentifier()
		if tok == token.ERR {
			return false
		}
		ll.emit(tok)

	fieldStuff:
		r := ll.next()
		switch r {
		case ' ', '\t':
			ll.acceptRun(" \t")
			ll.ignore()
			goto fieldStuff
		case '(':
			ll.emit(token.LPAREN)

			ok := ll.scanList(")", defListSep, 0, func(lll *lxr) bool {
				if lll.accept("\"") {
					sok := lll.scanString()
					if !sok {
						return false
					}
					lll.emit(token.DESCRIPTION)

					lll.ignoreSpace()
				}

				ident := lll.scanIdentifier()
				if ident == token.ERR {
					return false
				}
				lll.emit(ident)

				lll.acceptRun(" \t")
				lll.ignore()

				if !lll.accept(":") {
					ll.errorf("missing ':' in args definition")
					return false
				}

				vok := lll.scanVal()
				if !vok {
					return false
				}
				lll.acceptRun(" \t")
				lll.ignore()

				if lll.accept(",\n") {
					lll.backup()
					return true
				}

				if lll.accept("@") {
					lll.backup()
					return lll.scanDirectives(",)\r\n", " \t")
				}

				return true
			})
			if !ok {
				return false
			}
			ll.emit(token.RPAREN)

			ll.acceptRun(" \t")
			ll.ignore()

			if !ll.accept(":") {
				ll.errorf("missing ':' in fields definition")
				return false
			}

			fallthrough
		case ':':
			if !ll.scanVal() {
				return false
			}
			ll.acceptRun(" \t")
			ll.ignore()

			if !ll.accept("@") {
				break
			}

			fallthrough
		case '@':
			ll.backup()
			ok := ll.scanDirectives(",\r\n", " \t")
			if !ok {
				return false
			}
			ll.backup()
		default:
			ll.backup()
		}
		return true
	})
	if !ok {
		return nil
	}
	l.emit(token.RBRACE)

	return lexDoc
}

// lexUnion scans a union type definition
func lexUnion(l *lxr) stateFn {
	l.acceptRun(" \t")
	l.ignore()

	if l.accept("@") {
		l.backup()
		ok := l.scanDirectives("=\r\n", " \t")
		if !ok {
			return nil
		}
		l.backup()
	}

	if l.accept("\r\n") {
		return lexDoc
	}

	if l.accept("=") {
		l.emit(token.ASSIGN)
	}

	ok := l.scanList("\r\n", "|", '|', func(ll *lxr) bool {
		ident := ll.scanIdentifier()
		if ident == token.ERR {
			ll.errorf("invalid union member type identifier: %s", string(l.src[l.start:l.pos]))
			return false
		}
		ll.emit(ident)
		return true
	})
	if !ok {
		return nil
	}
	return lexDoc
}

// lexDirective scans a directive type definition
func lexDirective(l *lxr) stateFn {
	l.acceptRun(" \t")
	l.ignore()

	if !l.accept("@") {
		return l.errorf("directive decl must begin with a '@'")
	}
	l.emit(token.AT)

	name := l.scanIdentifier()
	if name == token.ERR {
		return l.errorf("invalid directive identifier: %s", string(l.src[l.start:l.pos]))
	}
	l.emit(name)

	l.acceptRun(" \t")
	l.ignore()

	if l.accept("(") {
		l.emit(token.LPAREN)

		ok := l.scanList(")", defListSep, 0, func(ll *lxr) bool {
			if ll.accept("\"") {
				sok := ll.scanString()
				if !sok {
					return false
				}
				ll.emit(token.DESCRIPTION)

				ll.ignoreSpace()
			}

			ident := ll.scanIdentifier()
			if ident == token.ERR {
				return false
			}
			ll.emit(ident)

			ll.acceptRun(" \t")
			ll.ignore()

			if !ll.accept(":") {
				ll.errorf("missing ':' in args definition")
				return false
			}

			vok := ll.scanVal()
			if !vok {
				return false
			}
			ll.acceptRun(" \t")
			ll.ignore()

			if ll.accept(",\n") {
				ll.backup()
				return true
			}

			if ll.accept("@") {
				ll.backup()
				return ll.scanDirectives(",)\r\n", " \t")
			}

			return true
		})
		if !ok {
			return nil
		}
		l.emit(token.RPAREN)
	}
	l.acceptRun(" \t")
	l.ignore()

	on := l.scanIdentifier()
	if on != token.ON {
		return l.errorf("directive decl must have locations specified with 'on' keyword, not: %s", string(l.src[l.start:l.pos]))
	}
	l.emit(token.ON)

	l.acceptRun(" \t")
	l.ignore()

	ok := l.scanList("\r\n", "|", '|', func(ll *lxr) bool {
		ident := ll.scanIdentifier()
		if ident == token.ERR {
			ll.errorf("invalid directive location identifier: %s", string(ll.src[ll.start:ll.pos]))
			return false
		}
		ll.emit(ident)
		return true
	})
	if !ok {
		return nil
	}
	return lexDoc
}

// scanVal scans an src value definitions, as used by obj, inter, and src types
func (l *lxr) scanVal() bool {
	l.emit(token.COLON)
	l.acceptRun(" \t")
	l.ignore()

	if l.accept("[") {
		l.emit(token.LBRACK)
	}

	typ := l.scanIdentifier()
	if typ == token.ERR {
		return false
	}
	l.emit(typ)

	if l.accept("!") {
		l.emit(token.NOT)
	}

	l.acceptRun(" \t")
	l.ignore()

	if l.accept("]") {
		l.emit(token.RBRACK)
	}

	if l.accept("!") {
		l.emit(token.NOT)
	}

	l.acceptRun(" \t")
	l.ignore()

	if l.accept("=") {
		l.emit(token.ASSIGN)
		l.acceptRun(" \t")
		l.ignore()

		ok := l.scanValue()
		if !ok {
			return false
		}
	}

	return true
}

// scanDirectives scans the list of directives on type defs
func (l *lxr) scanDirectives(endChars string, sep string) bool {
	return l.scanList(endChars, sep, 0, func(ll *lxr) bool {
		if !ll.accept("@") {
			l.errorf("directive must begin with an '@'")
			return false
		}
		ll.emit(token.AT)

		ident := ll.scanIdentifier()
		if ident == token.ERR {
			ll.errorf("invalid directive name: %s", ll.src[ll.start:ll.pos])
			return false
		}
		ll.emit(ident)

		r := ll.peek()
		switch r {
		case '(':
			break
		case '\n', '\r', ' ':
			return true
		}

		if !ll.accept("(") {
			return true
		}
		ll.emit(token.LPAREN)

		ok := ll.scanList(")", ",", ',', func(argsL *lxr) bool {
			id := argsL.scanIdentifier()
			if id == token.ERR {
				argsL.errorf("invalid argument name: %s", argsL.src[argsL.start:argsL.pos])
				return false
			}
			argsL.emit(id)

			argsL.acceptRun(" \t")
			argsL.ignore()

			if !argsL.accept(":") {
				argsL.errorf("expected ':' instead of: %s", string(argsL.src[argsL.pos]))
				return false
			}
			argsL.emit(token.COLON)

			argsL.acceptRun(" \t")
			argsL.ignore()

			return argsL.scanValue()
		})

		if ok {
			ll.emit(token.RPAREN)
		}
		return ok
	})
}

// scanValue scans a Value
func (l *lxr) scanValue() (ok bool) {
	var emitter func()

	switch r := l.peek(); {
	case r == '$':
		l.accept("$")
		l.emit(token.VAR)
		tok := l.scanIdentifier()
		if tok == token.ERR {
			emitter = func() { l.errorf("") }
		}
		emitter = func() { l.emit(tok) }
		ok = true
	case r == '"':
		ok = l.scanString()
		if !ok {
			emitter = func() { l.errorf("") }
		}
		emitter = func() { l.emit(token.STRING) }
	case isAlphaNumeric(r):
		if unicode.IsDigit(r) {
			num := l.scanNumber()
			if num == token.ERR {
				emitter = func() { l.errorf("") }
				break
			}
			emitter = func() { l.emit(num) }
			ok = true
			break
		}
		tok := l.scanIdentifier()
		if tok == token.ERR {
			emitter = func() { l.errorf("") }
			break
		}
		emitter = func() { l.emit(tok) }
		ok = true
	case r == '-':
		num := l.scanNumber()
		if num == token.ERR {
			emitter = func() { l.errorf("") }
			break
		}
		emitter = func() { l.emit(num) }
		ok = true
	case r == '[':
		l.accept("[")
		l.emit(token.LBRACK)
		ok = l.scanList("]", defListSep, 0, func(ll *lxr) bool {
			return l.scanValue()
		})
	case r == '{':
		l.accept("{")
		l.emit(token.LBRACE)
		ok = l.scanList("}", defListSep, 0, func(ll *lxr) bool {
			tok := ll.scanIdentifier()
			if tok == token.ERR {
				ll.errorf("invalid object field name: %s", ll.src[l.start:l.pos])
				return false
			}
			ll.emit(tok)

			ll.acceptRun(" \t")
			ll.ignore()

			if !ll.accept(":") {
				ll.errorf("expected field name-value seperator ':' but got %s", string(ll.src[l.pos]))
				return false
			}
			ll.emit(token.COLON)

			ll.acceptRun(" \t")
			ll.ignore()

			return ll.scanValue()
		})
	}

	emitter()

	return
}

const defListSep = ",\n"

// scanList scans a GraphQL list given a list element scanner func.
// The element scanner should assume that the lexer is right before whatever it needs
// scanList does not handle descriptions but it does handle comments
func (l *lxr) scanList(endDelims, sep string, rsep rune, elemScanner func(l *lxr) bool) bool {
	// start delim has already been lexed
	l.ignoreSpace()
	if r := l.next(); strings.ContainsRune(endDelims, r) {
		// return early if there was nothing in the list
		return true
	}
	l.backup()

	// Check if we hit comment
	if l.accept("#") {
		if l.mode != ScanComments {
			l.ignoreComment()
		} else {
			for r := l.next(); r != '\r' && r != '\n' && r != eof; {
				r = l.next()
			}
			l.emit(token.COMMENT)
		}
		return l.scanList(endDelims, sep, rsep, elemScanner)
	}

	// scan an element and return early if it failed
	ok := elemScanner(l)
	if !ok {
		return false
	}

	// Enforce same list seperator
loop:
	r := l.next()
	switch {
	case r == rsep:
		l.ignore()
	case r == ' ', r == '\t':
		if strings.ContainsRune(sep, r) {
			l.acceptRun(" \t")
			l.ignore()
			break
		}
		l.ignore()
		goto loop
	case r == '#':
		if sep == "," {
			l.errorf("expected a comma list seperator before comment in list")
			return false
		}
		l.ignoreComment()
	case strings.ContainsRune(endDelims, r):
		return true
	case strings.ContainsRune(sep, r):
		if rsep != 0 {
			if r == '\n' {
				l.backup()
			}
			l.errorf("list seperator must remain the same throughout the list")
			return false
		}

		l.ignore()
		rsep = r
	case r == eof:
		if strings.ContainsRune(endDelims, '\r') || strings.ContainsRune(endDelims, '\n') {
			l.backup()
			return true
		}
		fallthrough
	default:
		l.errorf("invalid list seperator: %v", r)
		return false
	}

	return l.scanList(endDelims, sep, rsep, elemScanner)
}

// scanString scans both a block string, `"""` and a normal string `"`
func (l *lxr) scanString() bool {
	l.acceptRun("\"")
	diff := l.pos - l.start
	if diff != 1 && diff != 3 {
		return false
	}

	for r := l.next(); r != '"' && r != eof; {
		if r == eof {
			return false
		}
		r = l.next()
	}
	l.backup()
	p := l.pos
	l.acceptRun("\"")

	newDiff := l.pos - p

	if newDiff != diff {
		return false
	}
	return true
}

// scanNumber scans both an int and a float as defined by the GraphQL spec.
func (l *lxr) scanNumber() token.Token {
	l.accept("-")
	l.acceptRun("0123456789")

	if !l.accept(".") && !l.accept("eE") {
		return token.INT
	}

	l.acceptRun("0123456789")
	l.accept("eE")
	l.accept("+-")
	l.acceptRun("0123456789")

	return token.FLOAT
}

// scanIdentifier scans an identifier and returns its token
func (l *lxr) scanIdentifier() token.Token {
	for r := l.next(); isAlphaNumeric(r); {
		r = l.next()
	}

	l.backup()
	word := string(l.src[l.start:l.pos])
	if !l.atTerminator() {
		return token.ERR
	}

	if l.peek() == '.' {
		l.emit(token.Lookup(word))
		l.accept(".")
		l.emit(token.PERIOD)
		return l.scanIdentifier()
	}

	return token.Lookup(word)
}

func (l *lxr) atTerminator() bool {
	r := l.peek()
	if isSpace(r) {
		return true
	}

	switch r {
	case eof, '.', ',', ':', ')', '(', '!', ']':
		return true
	}
	return false
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
