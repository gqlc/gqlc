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
		ok := l.scanList(')', defListSep, func(ll *lxr) bool {
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

	}
	return lexDoc
}

// lexObject scans a object type definition
func lexObject(l *lxr) stateFn {
	return lexDoc
}

// lexInterface scans a interface type definition
func lexInterface(l *lxr) stateFn {
	return lexDoc
}

// lexUnion scans a union type definition
func lexUnion(l *lxr) stateFn {
	return lexDoc
}

// lexEnum scans a enum type definition
func lexEnum(l *lxr) stateFn {
	return lexDoc
}

// lexInput scans a input type definition
func lexInput(l *lxr) stateFn {
	return lexDoc
}

// lexDirective scans a directive type definition
func lexDirective(l *lxr) stateFn {
	return lexDoc
}

// lexScalar scans a scalar type definition
func lexExtension(l *lxr) stateFn {
	return lexDoc
}

// scanDirectives scans the list of directives on type defs
func (l *lxr) scanDirectives() {

}

const defListSep = ",\n"

// scanList scans a GraphQL list given a list element scanner func.
// The element scanner should assume that the lexer is right before whatever it needs
// scanList does not handle descriptions but it does handle comments
func (l *lxr) scanList(endDelim rune, sep string, elemScanner func(l *lxr) bool) bool {
	// start delim has already been lexed
	l.ignoreSpace()
	if l.accept(string(endDelim)) {
		// return early if there was nothing in the list
		return true
	}

	// Check if we hit comment
	if l.accept("#") {
		l.ignoreComment()
		return l.scanList(endDelim, sep, elemScanner)
	}

	// scan an element and return early if it failed
	ok := elemScanner(l)
	if !ok {
		return false
	}

	// Now check for sep
	l.acceptRun(" \t")

	// Enforce same list seperator
	r := l.next()
	switch r {
	case ',', '\n':
		if string(r) != sep && sep != defListSep {
			if r == '\n' {
				l.backup()
			}
			l.errorf("list seperator must remain the same throughout the list")
			return false
		} else {
			sep = string(r)
		}
	case '#':
		if sep == "," {
			l.errorf("expected a comma list seperator before comment in list")
			return false
		}
		l.ignoreComment()
	case endDelim:
		l.backup()
		return l.scanList(endDelim, sep, elemScanner)
	default:
		l.errorf("invalid list seperator: %v", r)
		return false
	}
	l.ignore()

	return l.scanList(endDelim, sep, elemScanner)
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

func (l *lxr) scanNumber() bool {
	return true
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
