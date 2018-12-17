package lexer

import (
	"github.com/Zaba505/gqlc/graphql/token"
	"testing"
)

func TestLexImports(t *testing.T) {

	t.Run("single", func(subT *testing.T) {

		subT.Run("perfect", func(triT *testing.T) {
			dset := token.NewDocSet()
			src := []byte(`import "hello"`)
			l := Lex(dset.AddDoc("", dset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.STRING, Line: 1, Pos: 8, Val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("spaces", func(triT *testing.T) {
			dset := token.NewDocSet()
			src := []byte(`import 		    	  	 "hello"`)
			l := Lex(dset.AddDoc("", dset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.STRING, Line: 1, Pos: 19, Val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("noParenOrQuote", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import hello"`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.ERR, Line: 1, Pos: 8, Val: `missing ( or " to begin import statement`},
			}...)

		})
	})

	t.Run("multiple", func(subT *testing.T) {

		subT.Run("singleLine", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import ("hello")`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 9, Val: `"hello"`},
				{Typ: token.RPAREN, Line: 1, Pos: 16, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWComma", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import ( "a", "b", "c" )`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 10, Val: `"a"`},
				{Typ: token.STRING, Line: 1, Pos: 15, Val: `"b"`},
				{Typ: token.STRING, Line: 1, Pos: 20, Val: `"c"`},
				{Typ: token.RPAREN, Line: 1, Pos: 24, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWCommaNoEnd", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import ( "a", "b", "c" `)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 10, Val: `"a"`},
				{Typ: token.STRING, Line: 1, Pos: 15, Val: `"b"`},
				{Typ: token.STRING, Line: 1, Pos: 20, Val: `"c"`},
				{Typ: token.ERR, Line: 1, Pos: 24, Val: `invalid list seperator: -1`},
			}...)
		})

		subT.Run("multiLinesWNewLine", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import (
					"a"
					"b"
					"c"
				)`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 15, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 24, Val: `"b"`},
				{Typ: token.STRING, Line: 4, Pos: 33, Val: `"c"`},
				{Typ: token.RPAREN, Line: 5, Pos: 41, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("multiLinesWDiffSep1", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import (
					"a"
					"b",
					"c"
				)`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 15, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 24, Val: `"b"`},
				{Typ: token.ERR, Line: 3, Pos: 27, Val: `list seperator must remain the same throughout the list`},
			}...)
		})

		subT.Run("multiLinesWDiffSep2", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`import (
					"a",
					"b"
					"c",
				)`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 1, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 8, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 15, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 25, Val: `"b"`},
				{Typ: token.ERR, Line: 3, Pos: 28, Val: `list seperator must remain the same throughout the list`},
			}...)
		})

	})
}

func TestLexScalar(t *testing.T) {

	t.Run("simple", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`scalar URI`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
		expectItems(subT, l, []Item{
			{Typ: token.SCALAR, Line: 1, Pos: 1, Val: "scalar"},
			{Typ: token.IDENT, Line: 1, Pos: 8, Val: "URI"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`scalar URI @gotype @jstype() @darttype(if: Boolean)`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
		expectItems(subT, l, []Item{
			{Typ: token.SCALAR, Line: 1, Pos: 1, Val: "scalar"},
			{Typ: token.IDENT, Line: 1, Pos: 8, Val: "URI"},
			{Typ: token.AT, Line: 1, Pos: 12, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 13, Val: "gotype"},
			{Typ: token.AT, Line: 1, Pos: 20, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 21, Val: "jstype"},
			{Typ: token.LPAREN, Line: 1, Pos: 27, Val: "("},
			{Typ: token.RPAREN, Line: 1, Pos: 28, Val: ")"},
			{Typ: token.AT, Line: 1, Pos: 30, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 31, Val: "darttype"},
			{Typ: token.LPAREN, Line: 1, Pos: 39, Val: "("},
			{Typ: token.IDENT, Line: 1, Pos: 40, Val: "if"},
			{Typ: token.COLON, Line: 1, Pos: 42, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 44, Val: "Boolean"},
			{Typ: token.RPAREN, Line: 1, Pos: 51, Val: ")"},
		}...)
		expectEOF(subT, l)
	})
}

func TestScanValue(t *testing.T) {
	// Note:
	//    - Boolean, Null, and Enum values are all IDENT tokens so
	// 		they don't need to be tested here
	//	  - Strings are tested by LexImports
	//    - List and Objects are handled by ScanList which is implemented in others

	t.Run("var", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`$a`)
		l := &lxr{
			line:  1,
			items: make(chan Item),
			src:   src,
			doc:   fset.AddDoc("", fset.Base(), len(src)),
		}

		go func() {
			l.scanValue()
			close(l.items)
		}()

		expectItems(subT, l, []Item{
			{Typ: token.VAR, Line: 1, Pos: 1, Val: "$"},
			{Typ: token.IDENT, Line: 1, Pos: 2, Val: "a"},
		}...)
	})

	t.Run("int", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`12354654684013246813216513213254686210`)
		l := &lxr{
			line:  1,
			items: make(chan Item),
			src:   src,
			doc:   fset.AddDoc("", fset.Base(), len(src)),
		}

		go func() {
			l.scanValue()
			close(l.items)
		}()

		expectItems(subT, l,
			Item{Typ: token.INT, Line: 1, Pos: 1, Val: "12354654684013246813216513213254686210"},
		)
	})

	t.Run("float", func(subT *testing.T) {

		subT.Run("fractional", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`123.45`)
			l := &lxr{
				line:  1,
				items: make(chan Item),
				src:   src,
				doc:   fset.AddDoc("", fset.Base(), len(src)),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 1, Val: "123.45"},
			)
		})

		subT.Run("exponential", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`123e45`)
			l := &lxr{
				line:  1,
				items: make(chan Item),
				src:   src,
				doc:   fset.AddDoc("", fset.Base(), len(src)),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 1, Val: "123e45"},
			)
		})

		subT.Run("full", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`123.45e6`)
			l := &lxr{
				line:  1,
				items: make(chan Item),
				src:   src,
				doc:   fset.AddDoc("", fset.Base(), len(src)),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 1, Val: "123.45e6"},
			)
		})
	})
}

func TestLexObject(t *testing.T) {

	t.Run("withImpls", func(subT *testing.T) {

		subT.Run("perfect", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect implements One & Two & Three`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 28, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 34, Val: "Three"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("invalidSeperator", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect implements One , Two & Three`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.ERR, Line: 1, Pos: 26, Val: "invalid list seperator: 44"},
			}...)

			fset = token.NewDocSet()
			src = []byte(`type Rect implements One & Two , Three`)
			l = Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 28, Val: "Two"},
				{Typ: token.ERR, Line: 1, Pos: 32, Val: "invalid list seperator: 44"},
			}...)
		})
	})

	t.Run("withDirectives", func(subT *testing.T) {

		subT.Run("endsWithBrace", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect @green @blue {}`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 12, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 18, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 19, Val: "blue"},
				{Typ: token.LBRACE, Line: 1, Pos: 24, Val: "{"},
				{Typ: token.RBRACE, Line: 1, Pos: 25, Val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect @green @blue
`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 12, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 18, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 19, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect @green @blue`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 12, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 18, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 19, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})
	})

	t.Run("withImpls&Directives", func(subT *testing.T) {

		subT.Run("endsWithBrace", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect implements One & Two & Three @green @blue {}`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 28, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 34, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 40, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 41, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 47, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 48, Val: "blue"},
				{Typ: token.LBRACE, Line: 1, Pos: 53, Val: "{"},
				{Typ: token.RBRACE, Line: 1, Pos: 54, Val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect implements One & Two & Three @green @blue
`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 28, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 34, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 40, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 41, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 47, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 48, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			fset := token.NewDocSet()
			src := []byte(`type Rect implements One & Two & Three @green @blue`)
			l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 22, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 28, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 34, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 40, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 41, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 47, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 48, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})
	})

	t.Run("withFields", func(subT *testing.T) {

		subT.Run("asFieldsDef", func(triT *testing.T) {

			triT.Run("simple", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`type Rect {
	one: One
	two: Two
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 17, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 19, Val: "One"},
					{Typ: token.IDENT, Line: 3, Pos: 24, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 27, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 29, Val: "Two"},
					{Typ: token.RBRACE, Line: 4, Pos: 33, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDescrs", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`type Rect {
	"one descr" one: One
	"""
	two descr
	"""
	two: Two
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.DESCRIPTION, Line: 2, Pos: 14, Val: `"one descr"`},
					{Typ: token.IDENT, Line: 2, Pos: 26, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 29, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 31, Val: "One"},
					{Typ: token.DESCRIPTION, Line: 5, Pos: 36, Val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 6, Pos: 57, Val: "two"},
					{Typ: token.COLON, Line: 6, Pos: 60, Val: ":"},
					{Typ: token.IDENT, Line: 6, Pos: 62, Val: "Two"},
					{Typ: token.RBRACE, Line: 7, Pos: 66, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withArgs", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`type Rect {
	one(a: A, b: B): One
	two(
	"a descr" a: A
	"""
	b descr
	"""
	b: B
): Two
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "one"},
					{Typ: token.LPAREN, Line: 2, Pos: 17, Val: "("},
					{Typ: token.IDENT, Line: 2, Pos: 18, Val: "a"},
					{Typ: token.COLON, Line: 2, Pos: 19, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 21, Val: "A"},
					{Typ: token.IDENT, Line: 2, Pos: 24, Val: "b"},
					{Typ: token.COLON, Line: 2, Pos: 25, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 27, Val: "B"},
					{Typ: token.RPAREN, Line: 2, Pos: 28, Val: ")"},
					{Typ: token.COLON, Line: 2, Pos: 29, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 31, Val: "One"},
					{Typ: token.IDENT, Line: 3, Pos: 36, Val: "two"},
					{Typ: token.LPAREN, Line: 3, Pos: 39, Val: "("},
					{Typ: token.DESCRIPTION, Line: 4, Pos: 42, Val: `"a descr"`},
					{Typ: token.IDENT, Line: 4, Pos: 52, Val: "a"},
					{Typ: token.COLON, Line: 4, Pos: 53, Val: ":"},
					{Typ: token.IDENT, Line: 4, Pos: 55, Val: "A"},
					{Typ: token.DESCRIPTION, Line: 7, Pos: 58, Val: "\"\"\"\n\tb descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 8, Pos: 77, Val: "b"},
					{Typ: token.COLON, Line: 8, Pos: 78, Val: ":"},
					{Typ: token.IDENT, Line: 8, Pos: 80, Val: "B"},
					{Typ: token.RPAREN, Line: 9, Pos: 82, Val: ")"},
					{Typ: token.COLON, Line: 9, Pos: 83, Val: ":"},
					{Typ: token.IDENT, Line: 9, Pos: 85, Val: "Two"},
					{Typ: token.RBRACE, Line: 10, Pos: 89, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`type Rect {
	one: One @green @blue
	two: Two @blue
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 17, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 19, Val: "One"},
					{Typ: token.AT, Line: 2, Pos: 23, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 24, Val: "green"},
					{Typ: token.AT, Line: 2, Pos: 30, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 31, Val: "blue"},
					{Typ: token.IDENT, Line: 3, Pos: 37, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 40, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 42, Val: "Two"},
					{Typ: token.AT, Line: 3, Pos: 46, Val: "@"},
					{Typ: token.IDENT, Line: 3, Pos: 47, Val: "blue"},
					{Typ: token.RBRACE, Line: 4, Pos: 52, Val: "}"},
				}...)
				expectEOF(qt, l)
			})
		})

		subT.Run("asEnumValsDef", func(triT *testing.T) {

			triT.Run("simple", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`enum Rect {
	LEFT
	UP
	RIGHT
	DOWN
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 1, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "LEFT"},
					{Typ: token.IDENT, Line: 3, Pos: 20, Val: "UP"},
					{Typ: token.IDENT, Line: 4, Pos: 24, Val: "RIGHT"},
					{Typ: token.IDENT, Line: 5, Pos: 31, Val: "DOWN"},
					{Typ: token.RBRACE, Line: 6, Pos: 36, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDescrs", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`enum Rect {
	"left descr" LEFT
	"up descr" UP
	"""
	right descr
	"""
	RIGHT
	"down descr"
	DOWN
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 1, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.DESCRIPTION, Line: 2, Pos: 14, Val: `"left descr"`},
					{Typ: token.IDENT, Line: 2, Pos: 27, Val: "LEFT"},
					{Typ: token.DESCRIPTION, Line: 3, Pos: 33, Val: `"up descr"`},
					{Typ: token.IDENT, Line: 3, Pos: 44, Val: "UP"},
					{Typ: token.DESCRIPTION, Line: 6, Pos: 48, Val: "\"\"\"\n\tright descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 7, Pos: 71, Val: "RIGHT"},
					{Typ: token.DESCRIPTION, Line: 8, Pos: 78, Val: `"down descr"`},
					{Typ: token.IDENT, Line: 9, Pos: 92, Val: "DOWN"},
					{Typ: token.RBRACE, Line: 10, Pos: 97, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`enum Rect {
	LEFT @green @blue
	UP @red
	RIGHT
	DOWN @red @green @blue
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 1, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "LEFT"},
					{Typ: token.AT, Line: 2, Pos: 19, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 20, Val: "green"},
					{Typ: token.AT, Line: 2, Pos: 26, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 27, Val: "blue"},
					{Typ: token.IDENT, Line: 3, Pos: 33, Val: "UP"},
					{Typ: token.AT, Line: 3, Pos: 36, Val: "@"},
					{Typ: token.IDENT, Line: 3, Pos: 37, Val: "red"},
					{Typ: token.IDENT, Line: 4, Pos: 42, Val: "RIGHT"},
					{Typ: token.IDENT, Line: 5, Pos: 49, Val: "DOWN"},
					{Typ: token.AT, Line: 5, Pos: 54, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 55, Val: "red"},
					{Typ: token.AT, Line: 5, Pos: 59, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 60, Val: "green"},
					{Typ: token.AT, Line: 5, Pos: 66, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 67, Val: "blue"},
					{Typ: token.RBRACE, Line: 6, Pos: 72, Val: "}"},
				}...)
				expectEOF(qt, l)
			})
		})

		subT.Run("asInputFieldsDef", func(triT *testing.T) {

			triT.Run("simple", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`input Rect {
	one: One
	two: Two
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 1, Val: "input"},
					{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 12, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 15, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 18, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 20, Val: "One"},
					{Typ: token.IDENT, Line: 3, Pos: 25, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 28, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 30, Val: "Two"},
					{Typ: token.RBRACE, Line: 4, Pos: 34, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDescrs", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`input Rect {
	"one descr" one: One
	"""
	two descr
	"""
	two: Two
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 1, Val: "input"},
					{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 12, Val: "{"},
					{Typ: token.DESCRIPTION, Line: 2, Pos: 15, Val: `"one descr"`},
					{Typ: token.IDENT, Line: 2, Pos: 27, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 30, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 32, Val: "One"},
					{Typ: token.DESCRIPTION, Line: 5, Pos: 37, Val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 6, Pos: 58, Val: "two"},
					{Typ: token.COLON, Line: 6, Pos: 61, Val: ":"},
					{Typ: token.IDENT, Line: 6, Pos: 63, Val: "Two"},
					{Typ: token.RBRACE, Line: 7, Pos: 67, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDefVal", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`input Rect {
	one: One = 123
	two: Two = "abc"
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 1, Val: "input"},
					{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 12, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 15, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 18, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 20, Val: "One"},
					{Typ: token.ASSIGN, Line: 2, Pos: 24, Val: "="},
					{Typ: token.INT, Line: 2, Pos: 26, Val: "123"},
					{Typ: token.IDENT, Line: 3, Pos: 31, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 34, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 36, Val: "Two"},
					{Typ: token.ASSIGN, Line: 3, Pos: 40, Val: "="},
					{Typ: token.STRING, Line: 3, Pos: 42, Val: `"abc"`},
					{Typ: token.RBRACE, Line: 4, Pos: 48, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				fset := token.NewDocSet()
				src := []byte(`input Rect {
	one: One @green @blue
	two: Two @blue
}`)
				l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 1, Val: "input"},
					{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 12, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 15, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 18, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 20, Val: "One"},
					{Typ: token.AT, Line: 2, Pos: 24, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 25, Val: "green"},
					{Typ: token.AT, Line: 2, Pos: 31, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 32, Val: "blue"},
					{Typ: token.IDENT, Line: 3, Pos: 38, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 41, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 43, Val: "Two"},
					{Typ: token.AT, Line: 3, Pos: 47, Val: "@"},
					{Typ: token.IDENT, Line: 3, Pos: 48, Val: "blue"},
					{Typ: token.RBRACE, Line: 4, Pos: 53, Val: "}"},
				}...)
				expectEOF(qt, l)
			})
		})
	})

	t.Run("all", func(subT *testing.T) {
		// Note: This test does not use a valid GraphQL type decl.
		// 		 Instead, it uses a construction that is valid by the lexer and tests
		//		 the full capabilities of the lexObject stateFn.

		fset := token.NewDocSet()
		src := []byte(`type Rect implements Shape & Obj @green @blue {
	"one descr" one: One @one
	"""
	two descr
	"""
	two(
		"a descr" a: A = 1 @ptle
	): Two
	thr: Thr = 3 @ptle @ptle
}`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
		expectItems(subT, l, []Item{
			{Typ: token.TYPE, Line: 1, Pos: 1, Val: "type"},
			{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
			{Typ: token.IMPLEMENTS, Line: 1, Pos: 11, Val: "implements"},
			{Typ: token.IDENT, Line: 1, Pos: 22, Val: "Shape"},
			{Typ: token.IDENT, Line: 1, Pos: 30, Val: "Obj"},
			{Typ: token.AT, Line: 1, Pos: 34, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 35, Val: "green"},
			{Typ: token.AT, Line: 1, Pos: 41, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 42, Val: "blue"},
			{Typ: token.LBRACE, Line: 1, Pos: 47, Val: "{"},
			{Typ: token.DESCRIPTION, Line: 2, Pos: 50, Val: `"one descr"`},
			{Typ: token.IDENT, Line: 2, Pos: 62, Val: "one"},
			{Typ: token.COLON, Line: 2, Pos: 65, Val: ":"},
			{Typ: token.IDENT, Line: 2, Pos: 67, Val: "One"},
			{Typ: token.AT, Line: 2, Pos: 71, Val: "@"},
			{Typ: token.IDENT, Line: 2, Pos: 72, Val: "one"},
			{Typ: token.DESCRIPTION, Line: 5, Pos: 77, Val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
			{Typ: token.IDENT, Line: 6, Pos: 98, Val: "two"},
			{Typ: token.LPAREN, Line: 6, Pos: 101, Val: "("},
			{Typ: token.DESCRIPTION, Line: 7, Pos: 105, Val: `"a descr"`},
			{Typ: token.IDENT, Line: 7, Pos: 115, Val: "a"},
			{Typ: token.COLON, Line: 7, Pos: 116, Val: ":"},
			{Typ: token.IDENT, Line: 7, Pos: 118, Val: "A"},
			{Typ: token.ASSIGN, Line: 7, Pos: 120, Val: "="},
			{Typ: token.INT, Line: 7, Pos: 122, Val: "1"},
			{Typ: token.AT, Line: 7, Pos: 124, Val: "@"},
			{Typ: token.IDENT, Line: 7, Pos: 125, Val: "ptle"},
			{Typ: token.RPAREN, Line: 8, Pos: 131, Val: ")"},
			{Typ: token.COLON, Line: 8, Pos: 132, Val: ":"},
			{Typ: token.IDENT, Line: 8, Pos: 134, Val: "Two"},
			{Typ: token.IDENT, Line: 9, Pos: 139, Val: "thr"},
			{Typ: token.COLON, Line: 9, Pos: 142, Val: ":"},
			{Typ: token.IDENT, Line: 9, Pos: 144, Val: "Thr"},
			{Typ: token.ASSIGN, Line: 9, Pos: 148, Val: "="},
			{Typ: token.INT, Line: 9, Pos: 150, Val: "3"},
			{Typ: token.AT, Line: 9, Pos: 152, Val: "@"},
			{Typ: token.IDENT, Line: 9, Pos: 153, Val: "ptle"},
			{Typ: token.AT, Line: 9, Pos: 158, Val: "@"},
			{Typ: token.IDENT, Line: 9, Pos: 159, Val: "ptle"},
			{Typ: token.RBRACE, Line: 10, Pos: 164, Val: "}"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexUnion(t *testing.T) {

	t.Run("simple", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`union Pizza = Triangle | Circle`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
		expectItems(subT, l, []Item{
			{Typ: token.UNION, Line: 1, Pos: 1, Val: "union"},
			{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Pizza"},
			{Typ: token.ASSIGN, Line: 1, Pos: 13, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 15, Val: "Triangle"},
			{Typ: token.IDENT, Line: 1, Pos: 26, Val: "Circle"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`union Pizza @ham @pineapple = Triangle | Circle`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)
		expectItems(subT, l, []Item{
			{Typ: token.UNION, Line: 1, Pos: 1, Val: "union"},
			{Typ: token.IDENT, Line: 1, Pos: 7, Val: "Pizza"},
			{Typ: token.AT, Line: 1, Pos: 13, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 14, Val: "ham"},
			{Typ: token.AT, Line: 1, Pos: 18, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 19, Val: "pineapple"},
			{Typ: token.ASSIGN, Line: 1, Pos: 29, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 31, Val: "Triangle"},
			{Typ: token.IDENT, Line: 1, Pos: 42, Val: "Circle"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexDirective(t *testing.T) {

	t.Run("simple", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`directive @skip on FIELD | FIELD_DEFINITION`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)

		expectItems(subT, l, []Item{
			{Typ: token.DIRECTIVE, Line: 1, Pos: 1, Val: "directive"},
			{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 12, Val: "skip"},
			{Typ: token.ON, Line: 1, Pos: 17, Val: "on"},
			{Typ: token.IDENT, Line: 1, Pos: 20, Val: "FIELD"},
			{Typ: token.IDENT, Line: 1, Pos: 28, Val: "FIELD_DEFINITION"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withArgs", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`directive @skip(if: Boolean, else: Boolean = false) on FIELD | FIELD_DEFINITION`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)

		expectItems(subT, l, []Item{
			{Typ: token.DIRECTIVE, Line: 1, Pos: 1, Val: "directive"},
			{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 12, Val: "skip"},
			{Typ: token.LPAREN, Line: 1, Pos: 16, Val: "("},
			{Typ: token.IDENT, Line: 1, Pos: 17, Val: "if"},
			{Typ: token.COLON, Line: 1, Pos: 19, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 21, Val: "Boolean"},
			{Typ: token.IDENT, Line: 1, Pos: 30, Val: "else"},
			{Typ: token.COLON, Line: 1, Pos: 34, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 36, Val: "Boolean"},
			{Typ: token.ASSIGN, Line: 1, Pos: 44, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 46, Val: "false"},
			{Typ: token.RPAREN, Line: 1, Pos: 51, Val: ")"},
			{Typ: token.ON, Line: 1, Pos: 53, Val: "on"},
			{Typ: token.IDENT, Line: 1, Pos: 56, Val: "FIELD"},
			{Typ: token.IDENT, Line: 1, Pos: 64, Val: "FIELD_DEFINITION"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("argsWithDirectives", func(subT *testing.T) {
		fset := token.NewDocSet()
		src := []byte(`directive @skip(if: Boolean @one(), else: Boolean = false @one() @two()) on FIELD | FIELD_DEFINITION`)
		l := Lex(fset.AddDoc("", fset.Base(), len(src)), src, 0)

		expectItems(subT, l, []Item{
			{Typ: token.DIRECTIVE, Line: 1, Pos: 1, Val: "directive"},
			{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 12, Val: "skip"},
			{Typ: token.LPAREN, Line: 1, Pos: 16, Val: "("},
			{Typ: token.IDENT, Line: 1, Pos: 17, Val: "if"},
			{Typ: token.COLON, Line: 1, Pos: 19, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 21, Val: "Boolean"},
			{Typ: token.AT, Line: 1, Pos: 29, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 30, Val: "one"},
			{Typ: token.LPAREN, Line: 1, Pos: 33, Val: "("},
			{Typ: token.RPAREN, Line: 1, Pos: 34, Val: ")"},
			{Typ: token.IDENT, Line: 1, Pos: 37, Val: "else"},
			{Typ: token.COLON, Line: 1, Pos: 41, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 43, Val: "Boolean"},
			{Typ: token.ASSIGN, Line: 1, Pos: 51, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 53, Val: "false"},
			{Typ: token.AT, Line: 1, Pos: 59, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 60, Val: "one"},
			{Typ: token.LPAREN, Line: 1, Pos: 63, Val: "("},
			{Typ: token.RPAREN, Line: 1, Pos: 64, Val: ")"},
			{Typ: token.AT, Line: 1, Pos: 66, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 67, Val: "two"},
			{Typ: token.LPAREN, Line: 1, Pos: 70, Val: "("},
			{Typ: token.RPAREN, Line: 1, Pos: 71, Val: ")"},
			{Typ: token.RPAREN, Line: 1, Pos: 72, Val: ")"},
			{Typ: token.ON, Line: 1, Pos: 74, Val: "on"},
			{Typ: token.IDENT, Line: 1, Pos: 77, Val: "FIELD"},
			{Typ: token.IDENT, Line: 1, Pos: 85, Val: "FIELD_DEFINITION"},
		}...)
		expectEOF(subT, l)
	})
}

func expectItems(t *testing.T, l Interface, items ...Item) {
	for _, item := range items {
		lItem := l.NextItem()
		if lItem != item {
			t.Fatalf("expected item: %#v but instead received: %#v", item, lItem)
		}
	}
}

func expectEOF(t *testing.T, l Interface) {
	i := l.NextItem()
	if i.Typ != token.EOF {
		t.Fatalf("expected eof but instead received: %#v", i)
	}
}
