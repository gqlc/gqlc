package lexer

import (
	"fmt"
	"gqlc/sl/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Item struct {
	typ  token.Token
	pos  token.Pos
	val  string
	line int
}

func (i Item) String() string {
	switch {
	case i.typ == token.EOF:
		return "EOF"
	case i.typ == token.ERR:
		return i.val
	case i.typ >= token.PACKAGE:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type Interface interface {
	NextItem() Item
	Drain()
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lxr) stateFn

type lxr struct {
	name  string
	input string
	pos   token.Pos
	start token.Pos
	width token.Pos
	items chan Item
	line  int
}

func Lex(name, input string) Interface {
	l := &lxr{
		name:  name,
		input: input,
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

// next returns the next rune in the input.
func (l *lxr) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = token.Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lxr) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lxr) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

// emit passes an item back to the client.
func (l *lxr) emit(t token.Token) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos], l.line}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
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
	l.items <- Item{token.ERR, l.start, fmt.Sprintf(format, args...), l.line}
	return nil
}

// ignoreSpace consume all whitespace
func (l *lxr) ignoreSpace() {
	l.acceptRun(spaceChars)
	l.ignore()
}

// NextItem returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lxr) NextItem() Item {
	return <-l.items
}

// drain drains the output so the lexing goroutine will exit.
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
		l.ignoreComment()
	case r == '"':
		l.backup()
		if !l.scanString() {
			return l.errorf("bad string syntax: %q", l.input[l.start:l.pos])
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

	switch ident {
	case token.IMPORT:
		return lexImports
	case token.SCALAR:
		return lexScalar
	case token.TYPE:
		return lexObject
	case token.INTERFACE:
		return lexInterface
	case token.UNION:
		return lexUnion
	case token.ENUM:
		return lexEnum
	case token.INPUT:
		return lexInput
	case token.DIRECTIVE:
		return lexDirective
	case token.EXTEND:
		return lexExtension
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
			return l.errorf("malformed input string: %s", l.input[l.start:l.pos])
		}
		l.emit(token.STRING)
	case '(':
		l.accept("(")
		l.emit(token.LPAREN)
		ok := l.scanList(")", defListSep, func(ll *lxr) bool {
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

	if !isAlphaNumeric(l.peek()) {
		return l.errorf("expected scalar type identifier name")
	}

	ident := l.scanIdentifier()
	l.emit(ident)

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

// lexObject scans a object type definition
func lexObject(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexInterface scans a interface type definition
func lexInterface(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexUnion scans a union type definition
func lexUnion(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexEnum scans a enum type definition
func lexEnum(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexInput scans a input type definition
func lexInput(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexDirective scans a directive type definition
func lexDirective(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// lexScalar scans a scalar type definition
func lexExtension(l *lxr) stateFn {
	// TODO
	return lexDoc
}

// scanDirectives scans the list of directives on type defs
func (l *lxr) scanDirectives(endChars string, sep string) bool {
	return l.scanList(endChars, sep, func(ll *lxr) bool {
		if !ll.accept("@") {
			l.errorf("directive must begin with an '@'")
			return false
		}
		ll.emit(token.AT)

		ident := ll.scanIdentifier()
		if ident == token.ERR {
			ll.errorf("invalid directive name: %s", ll.input[ll.start:ll.pos])
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

		ll.accept("(")
		ll.emit(token.LPAREN)

		ok := ll.scanList(")", ",", func(argsL *lxr) bool {
			id := argsL.scanIdentifier()
			if id == token.ERR {
				argsL.errorf("invalid argument name: %s", argsL.input[argsL.start:argsL.pos])
				return false
			}
			argsL.emit(id)

			argsL.acceptRun(" \t")
			argsL.ignore()

			if !argsL.accept(":") {
				argsL.errorf("expected ':' instead of: %s", string(argsL.input[argsL.pos]))
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
		ok = l.scanList("]", defListSep, func(ll *lxr) bool {
			return l.scanValue()
		})
	case r == '{':
		l.accept("{")
		l.emit(token.LBRACE)
		ok = l.scanList("}", defListSep, func(ll *lxr) bool {
			tok := ll.scanIdentifier()
			if tok == token.ERR {
				ll.errorf("invalid object field name: %s", ll.input[l.start:l.pos])
				return false
			}
			ll.emit(tok)

			ll.acceptRun(" \t")
			ll.ignore()

			if !ll.accept(":") {
				ll.errorf("expected field name-value seperator ':' but got %s", string(ll.input[l.pos]))
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
func (l *lxr) scanList(endDelims, sep string, elemScanner func(l *lxr) bool) bool {
	// start delim has already been lexed
	l.ignoreSpace()
	if r := l.next(); strings.ContainsRune(endDelims, r) {
		// return early if there was nothing in the list
		return true
	}
	l.backup()

	// Check if we hit comment
	if l.accept("#") {
		l.ignoreComment()
		return l.scanList(endDelims, sep, elemScanner)
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
	case r == ',', r == '\n', r == '\r':
		// Check if newline is list endDelim
		if strings.ContainsRune(endDelims, r) {
			l.ignore()
			return true
		}

		if string(r) != sep && sep != defListSep {
			if r == '\n' {
				l.backup()
			}
			l.errorf("list seperator must remain the same throughout the list")
			return false
		} else {
			sep = string(r)
		}
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
		l.backup()
		l.ignore()
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

	return l.scanList(endDelims, sep, elemScanner)
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
	word := l.input[l.start:l.pos]
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
	case eof, '.', ',', ':', ')', '(':
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
