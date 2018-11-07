package lexer

import (
	"gqlc/sl/token"
	"testing"
)

func TestLexImports(t *testing.T) {
	t.Run("single", func(subT *testing.T) {
		subT.Run("perfect", func(triT *testing.T) {
			l := Lex("perfect", []byte(`import "hello"`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Val: "import"},
				{Typ: token.STRING, Pos: 7, Line: 1, Val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("spaces", func(triT *testing.T) {
			l := Lex("spaces", []byte(`import 		    	  	 "hello"`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Val: "import"},
				{Typ: token.STRING, Pos: 18, Line: 1, Val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("noParenOrQuote", func(triT *testing.T) {
			l := Lex("noParenOrQuote", []byte(`import hello"`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Val: "import"},
				{Typ: token.ERR, Pos: 7, Line: 1, Val: `missing ( or " to begin import statement`},
			}...)

		})
	})

	t.Run("multiple", func(subT *testing.T) {
		subT.Run("singleLine", func(triT *testing.T) {
			l := Lex("singleLine", []byte(`import ("hello")`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 8, Val: `"hello"`},
				{Typ: token.RPAREN, Line: 1, Pos: 15, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWComma", func(triT *testing.T) {
			l := Lex("singleLineWComma", []byte(`import ( "a", "b", "c" )`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 9, Val: `"a"`},
				{Typ: token.STRING, Line: 1, Pos: 14, Val: `"b"`},
				{Typ: token.STRING, Line: 1, Pos: 19, Val: `"c"`},
				{Typ: token.RPAREN, Line: 1, Pos: 23, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWCommaNoEnd", func(triT *testing.T) {
			l := Lex("singleLineWCommaNoEnd", []byte(`import ( "a", "b", "c" `), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 1, Pos: 9, Val: `"a"`},
				{Typ: token.STRING, Line: 1, Pos: 14, Val: `"b"`},
				{Typ: token.STRING, Line: 1, Pos: 19, Val: `"c"`},
				{Typ: token.ERR, Line: 1, Pos: 23, Val: `invalid list seperator: -1`},
			}...)
		})

		subT.Run("multiLinesWNewLine", func(triT *testing.T) {
			l := Lex("multLinesWNewLine",
				[]byte(`import (
					"a"
					"b"
					"c"
				)`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 14, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 23, Val: `"b"`},
				{Typ: token.STRING, Line: 4, Pos: 32, Val: `"c"`},
				{Typ: token.RPAREN, Line: 5, Pos: 40, Val: ")"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("multiLinesWDiffSep1", func(triT *testing.T) {
			l := Lex("multLinesWDiffSep1",
				[]byte(`import (
					"a"
					"b",
					"c"
				)`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 14, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 23, Val: `"b"`},
				{Typ: token.ERR, Line: 3, Pos: 26, Val: `list seperator must remain the same throughout the list`},
			}...)
		})

		subT.Run("multiLinesWDiffSep2", func(triT *testing.T) {
			l := Lex("multLinesWDiffSep2",
				[]byte(`import (
					"a",
					"b"
					"c",
				)`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.IMPORT, Line: 1, Pos: 0, Val: "import"},
				{Typ: token.LPAREN, Line: 1, Pos: 7, Val: "("},
				{Typ: token.STRING, Line: 2, Pos: 14, Val: `"a"`},
				{Typ: token.STRING, Line: 3, Pos: 24, Val: `"b"`},
				{Typ: token.ERR, Line: 3, Pos: 27, Val: `list seperator must remain the same throughout the list`},
			}...)
		})

	})
}

func TestLexScalar(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`scalar URI`), 0)
		expectItems(subT, l, []Item{
			{Typ: token.SCALAR, Line: 1, Pos: 0, Val: "scalar"},
			{Typ: token.IDENT, Line: 1, Pos: 7, Val: "URI"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		l := Lex("withDirectives", []byte(`scalar URI @gotype @jstype() @darttype(if: Boolean)`), 0)
		expectItems(subT, l, []Item{
			{Typ: token.SCALAR, Line: 1, Pos: 0, Val: "scalar"},
			{Typ: token.IDENT, Line: 1, Pos: 7, Val: "URI"},
			{Typ: token.AT, Line: 1, Pos: 11, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 12, Val: "gotype"},
			{Typ: token.AT, Line: 1, Pos: 19, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 20, Val: "jstype"},
			{Typ: token.LPAREN, Line: 1, Pos: 26, Val: "("},
			{Typ: token.RPAREN, Line: 1, Pos: 27, Val: ")"},
			{Typ: token.AT, Line: 1, Pos: 29, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 30, Val: "darttype"},
			{Typ: token.LPAREN, Line: 1, Pos: 38, Val: "("},
			{Typ: token.IDENT, Line: 1, Pos: 39, Val: "if"},
			{Typ: token.COLON, Line: 1, Pos: 41, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 43, Val: "Boolean"},
			{Typ: token.RPAREN, Line: 1, Pos: 50, Val: ")"},
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
		l := &lxr{
			line:  1,
			items: make(chan Item),
			input: []byte(`$a`),
		}

		go func() {
			l.scanValue()
			close(l.items)
		}()

		expectItems(subT, l, []Item{
			{Typ: token.VAR, Line: 1, Pos: 0, Val: "$"},
			{Typ: token.IDENT, Line: 1, Pos: 1, Val: "a"},
		}...)
	})

	t.Run("int", func(subT *testing.T) {
		l := &lxr{
			line:  1,
			items: make(chan Item),
			input: []byte(`12354654684013246813216513213254686210`),
		}

		go func() {
			l.scanValue()
			close(l.items)
		}()

		expectItems(subT, l,
			Item{Typ: token.INT, Line: 1, Pos: 0, Val: "12354654684013246813216513213254686210"},
		)
	})

	t.Run("float", func(subT *testing.T) {
		subT.Run("fractional", func(triT *testing.T) {
			l := &lxr{
				line:  1,
				items: make(chan Item),
				input: []byte(`123.45`),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 0, Val: "123.45"},
			)
		})

		subT.Run("exponential", func(triT *testing.T) {
			l := &lxr{
				line:  1,
				items: make(chan Item),
				input: []byte(`123e45`),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 0, Val: "123e45"},
			)
		})

		subT.Run("full", func(triT *testing.T) {
			l := &lxr{
				line:  1,
				items: make(chan Item),
				input: []byte(`123.45e6`),
			}

			go func() {
				l.scanValue()
				close(l.items)
			}()

			expectItems(subT, l,
				Item{Typ: token.FLOAT, Line: 1, Pos: 0, Val: "123.45e6"},
			)
		})
	})
}

func TestLexObject(t *testing.T) {
	t.Run("withImpls", func(subT *testing.T) {

		subT.Run("perfect", func(triT *testing.T) {
			l := Lex("perfect", []byte(`type Rect implements One & Two & Three`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 27, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 33, Val: "Three"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("invalidSeperator", func(triT *testing.T) {
			l := Lex("first", []byte(`type Rect implements One , Two & Three`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.ERR, Line: 1, Pos: 25, Val: "invalid list seperator: 44"},
			}...)

			l = Lex("later", []byte(`type Rect implements One & Two , Three`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 27, Val: "Two"},
				{Typ: token.ERR, Line: 1, Pos: 31, Val: "invalid list seperator: 44"},
			}...)
		})
	})

	t.Run("withDirectives", func(subT *testing.T) {

		subT.Run("endsWithBrace", func(triT *testing.T) {
			l := Lex("endsWithBrace", []byte(`type Rect @green @blue {}`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 10, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 11, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 17, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 18, Val: "blue"},
				{Typ: token.LBRACE, Line: 1, Pos: 23, Val: "{"},
				{Typ: token.RBRACE, Line: 1, Pos: 24, Val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			l := Lex("endsWithNewline", []byte(`type Rect @green @blue
`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 10, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 11, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 17, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 18, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			l := Lex("endsWithEOF", []byte(`type Rect @green @blue`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.AT, Line: 1, Pos: 10, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 11, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 17, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 18, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})
	})

	t.Run("withImpls&Directives", func(subT *testing.T) {

		subT.Run("endsWithBrace", func(triT *testing.T) {
			l := Lex("endsWithBrace", []byte(`type Rect implements One & Two & Three @green @blue {}`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 27, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 33, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 39, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 40, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 46, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 47, Val: "blue"},
				{Typ: token.LBRACE, Line: 1, Pos: 52, Val: "{"},
				{Typ: token.RBRACE, Line: 1, Pos: 53, Val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			l := Lex("endsWithNewline", []byte(`type Rect implements One & Two & Three @green @blue
`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 27, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 33, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 39, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 40, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 46, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 47, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			l := Lex("endsWithEOF", []byte(`type Rect implements One & Two & Three @green @blue`), 0)
			expectItems(triT, l, []Item{
				{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
				{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
				{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
				{Typ: token.IDENT, Line: 1, Pos: 21, Val: "One"},
				{Typ: token.IDENT, Line: 1, Pos: 27, Val: "Two"},
				{Typ: token.IDENT, Line: 1, Pos: 33, Val: "Three"},
				{Typ: token.AT, Line: 1, Pos: 39, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 40, Val: "green"},
				{Typ: token.AT, Line: 1, Pos: 46, Val: "@"},
				{Typ: token.IDENT, Line: 1, Pos: 47, Val: "blue"},
			}...)
			expectEOF(triT, l)
		})
	})

	t.Run("withFields", func(subT *testing.T) {

		subT.Run("asFieldsDef", func(triT *testing.T) {

			triT.Run("simple", func(qt *testing.T) {
				l := Lex("simple", []byte(`type Rect {
	one: One
	two: Two
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 13, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 16, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 18, Val: "One"},
					{Typ: token.IDENT, Line: 3, Pos: 23, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 26, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 28, Val: "Two"},
					{Typ: token.RBRACE, Line: 4, Pos: 32, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDescrs", func(qt *testing.T) {
				l := Lex("withDescrs", []byte(`type Rect {
	"one descr" one: One
	"""
	two descr
	"""
	two: Two
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.DESCRIPTION, Line: 2, Pos: 13, Val: `"one descr"`},
					{Typ: token.IDENT, Line: 2, Pos: 25, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 28, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 30, Val: "One"},
					{Typ: token.DESCRIPTION, Line: 5, Pos: 35, Val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 6, Pos: 56, Val: "two"},
					{Typ: token.COLON, Line: 6, Pos: 59, Val: ":"},
					{Typ: token.IDENT, Line: 6, Pos: 61, Val: "Two"},
					{Typ: token.RBRACE, Line: 7, Pos: 65, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withArgs", func(qt *testing.T) {
				l := Lex("withArgs", []byte(`type Rect {
	one(a: A, b: B): One
	two(
	"a descr" a: A
	"""
	b descr
	"""
	b: B
): Two
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 13, Val: "one"},
					{Typ: token.LPAREN, Line: 2, Pos: 16, Val: "("},
					{Typ: token.IDENT, Line: 2, Pos: 17, Val: "a"},
					{Typ: token.COLON, Line: 2, Pos: 18, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 20, Val: "A"},
					{Typ: token.IDENT, Line: 2, Pos: 23, Val: "b"},
					{Typ: token.COLON, Line: 2, Pos: 24, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 26, Val: "B"},
					{Typ: token.RPAREN, Line: 2, Pos: 27, Val: ")"},
					{Typ: token.COLON, Line: 2, Pos: 28, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 30, Val: "One"},
					{Typ: token.IDENT, Line: 3, Pos: 35, Val: "two"},
					{Typ: token.LPAREN, Line: 3, Pos: 38, Val: "("},
					{Typ: token.DESCRIPTION, Line: 4, Pos: 41, Val: `"a descr"`},
					{Typ: token.IDENT, Line: 4, Pos: 51, Val: "a"},
					{Typ: token.COLON, Line: 4, Pos: 52, Val: ":"},
					{Typ: token.IDENT, Line: 4, Pos: 54, Val: "A"},
					{Typ: token.DESCRIPTION, Line: 7, Pos: 57, Val: "\"\"\"\n\tb descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 8, Pos: 76, Val: "b"},
					{Typ: token.COLON, Line: 8, Pos: 77, Val: ":"},
					{Typ: token.IDENT, Line: 8, Pos: 79, Val: "B"},
					{Typ: token.RPAREN, Line: 9, Pos: 81, Val: ")"},
					{Typ: token.COLON, Line: 9, Pos: 82, Val: ":"},
					{Typ: token.IDENT, Line: 9, Pos: 84, Val: "Two"},
					{Typ: token.RBRACE, Line: 10, Pos: 88, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				l := Lex("withDirectives", []byte(`type Rect {
	one: One @green @blue
	two: Two @blue
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 13, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 16, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 18, Val: "One"},
					{Typ: token.AT, Line: 2, Pos: 22, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 23, Val: "green"},
					{Typ: token.AT, Line: 2, Pos: 29, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 30, Val: "blue"},
					{Typ: token.IDENT, Line: 3, Pos: 36, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 39, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 41, Val: "Two"},
					{Typ: token.AT, Line: 3, Pos: 45, Val: "@"},
					{Typ: token.IDENT, Line: 3, Pos: 46, Val: "blue"},
					{Typ: token.RBRACE, Line: 4, Pos: 51, Val: "}"},
				}...)
				expectEOF(qt, l)
			})
		})

		subT.Run("asEnumValsDef", func(triT *testing.T) {

			triT.Run("simple", func(qt *testing.T) {
				l := Lex("simple", []byte(`enum Rect {
	LEFT
	UP
	RIGHT
	DOWN
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 0, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 13, Val: "LEFT"},
					{Typ: token.IDENT, Line: 3, Pos: 19, Val: "UP"},
					{Typ: token.IDENT, Line: 4, Pos: 23, Val: "RIGHT"},
					{Typ: token.IDENT, Line: 5, Pos: 30, Val: "DOWN"},
					{Typ: token.RBRACE, Line: 6, Pos: 35, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDescrs", func(qt *testing.T) {
				l := Lex("withDescrs", []byte(`enum Rect {
	"left descr" LEFT
	"up descr" UP
	"""
	right descr
	"""
	RIGHT
	"down descr"
	DOWN
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 0, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.DESCRIPTION, Line: 2, Pos: 13, Val: `"left descr"`},
					{Typ: token.IDENT, Line: 2, Pos: 26, Val: "LEFT"},
					{Typ: token.DESCRIPTION, Line: 3, Pos: 32, Val: `"up descr"`},
					{Typ: token.IDENT, Line: 3, Pos: 43, Val: "UP"},
					{Typ: token.DESCRIPTION, Line: 6, Pos: 47, Val: "\"\"\"\n\tright descr\n\t\"\"\""},
					{Typ: token.IDENT, Line: 7, Pos: 70, Val: "RIGHT"},
					{Typ: token.DESCRIPTION, Line: 8, Pos: 77, Val: `"down descr"`},
					{Typ: token.IDENT, Line: 9, Pos: 91, Val: "DOWN"},
					{Typ: token.RBRACE, Line: 10, Pos: 96, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				l := Lex("withDirectives", []byte(`enum Rect {
	LEFT @green @blue
	UP @red
	RIGHT
	DOWN @red @green @blue
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.ENUM, Line: 1, Pos: 0, Val: "enum"},
					{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 10, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 13, Val: "LEFT"},
					{Typ: token.AT, Line: 2, Pos: 18, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 19, Val: "green"},
					{Typ: token.AT, Line: 2, Pos: 25, Val: "@"},
					{Typ: token.IDENT, Line: 2, Pos: 26, Val: "blue"},
					{Typ: token.IDENT, Line: 3, Pos: 32, Val: "UP"},
					{Typ: token.AT, Line: 3, Pos: 35, Val: "@"},
					{Typ: token.IDENT, Line: 3, Pos: 36, Val: "red"},
					{Typ: token.IDENT, Line: 4, Pos: 41, Val: "RIGHT"},
					{Typ: token.IDENT, Line: 5, Pos: 48, Val: "DOWN"},
					{Typ: token.AT, Line: 5, Pos: 53, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 54, Val: "red"},
					{Typ: token.AT, Line: 5, Pos: 58, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 59, Val: "green"},
					{Typ: token.AT, Line: 5, Pos: 65, Val: "@"},
					{Typ: token.IDENT, Line: 5, Pos: 66, Val: "blue"},
					{Typ: token.RBRACE, Line: 6, Pos: 71, Val: "}"},
				}...)
				expectEOF(qt, l)
			})
		})

		subT.Run("asInputFieldsDef", func(triT *testing.T) {
			triT.Run("simple", func(qt *testing.T) {
				l := Lex("simple", []byte(`input Rect {
	one: One
	two: Two
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 0, Val: "input"},
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
				l := Lex("withDescrs", []byte(`input Rect {
	"one descr" one: One
	"""
	two descr
	"""
	two: Two
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 0, Val: "input"},
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

			triT.Run("withDefVal", func(qt *testing.T) {
				l := Lex("withDefVal", []byte(`input Rect {
	one: One = 123
	two: Two = "abc"
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 0, Val: "input"},
					{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Rect"},
					{Typ: token.LBRACE, Line: 1, Pos: 11, Val: "{"},
					{Typ: token.IDENT, Line: 2, Pos: 14, Val: "one"},
					{Typ: token.COLON, Line: 2, Pos: 17, Val: ":"},
					{Typ: token.IDENT, Line: 2, Pos: 19, Val: "One"},
					{Typ: token.ASSIGN, Line: 2, Pos: 23, Val: "="},
					{Typ: token.INT, Line: 2, Pos: 25, Val: "123"},
					{Typ: token.IDENT, Line: 3, Pos: 30, Val: "two"},
					{Typ: token.COLON, Line: 3, Pos: 33, Val: ":"},
					{Typ: token.IDENT, Line: 3, Pos: 35, Val: "Two"},
					{Typ: token.ASSIGN, Line: 3, Pos: 39, Val: "="},
					{Typ: token.STRING, Line: 3, Pos: 41, Val: `"abc"`},
					{Typ: token.RBRACE, Line: 4, Pos: 47, Val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				l := Lex("withDirectives", []byte(`input Rect {
	one: One @green @blue
	two: Two @blue
}`), 0)
				expectItems(qt, l, []Item{
					{Typ: token.INPUT, Line: 1, Pos: 0, Val: "input"},
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
	})

	t.Run("all", func(subT *testing.T) {
		// Note: This test does not use a valid GraphQL type decl.
		// 		 Instead, it uses a construction that is valid by the lexer and tests
		//		 the full capabilities of the lexObject stateFn.

		l := Lex("all", []byte(`type Rect implements Shape & Obj @green @blue {
	"one descr" one: One @one
	"""
	two descr
	"""
	two(
		"a descr" a: A = 1 @ptle
	): Two
	thr: Thr = 3 @ptle @ptle
}`), 0)
		expectItems(subT, l, []Item{
			{Typ: token.TYPE, Line: 1, Pos: 0, Val: "type"},
			{Typ: token.IDENT, Line: 1, Pos: 5, Val: "Rect"},
			{Typ: token.IMPLEMENTS, Line: 1, Pos: 10, Val: "implements"},
			{Typ: token.IDENT, Line: 1, Pos: 21, Val: "Shape"},
			{Typ: token.IDENT, Line: 1, Pos: 29, Val: "Obj"},
			{Typ: token.AT, Line: 1, Pos: 33, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 34, Val: "green"},
			{Typ: token.AT, Line: 1, Pos: 40, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 41, Val: "blue"},
			{Typ: token.LBRACE, Line: 1, Pos: 46, Val: "{"},
			{Typ: token.DESCRIPTION, Line: 2, Pos: 49, Val: `"one descr"`},
			{Typ: token.IDENT, Line: 2, Pos: 61, Val: "one"},
			{Typ: token.COLON, Line: 2, Pos: 64, Val: ":"},
			{Typ: token.IDENT, Line: 2, Pos: 66, Val: "One"},
			{Typ: token.AT, Line: 2, Pos: 70, Val: "@"},
			{Typ: token.IDENT, Line: 2, Pos: 71, Val: "one"},
			{Typ: token.DESCRIPTION, Line: 5, Pos: 76, Val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
			{Typ: token.IDENT, Line: 6, Pos: 97, Val: "two"},
			{Typ: token.LPAREN, Line: 6, Pos: 100, Val: "("},
			{Typ: token.DESCRIPTION, Line: 7, Pos: 104, Val: `"a descr"`},
			{Typ: token.IDENT, Line: 7, Pos: 114, Val: "a"},
			{Typ: token.COLON, Line: 7, Pos: 115, Val: ":"},
			{Typ: token.IDENT, Line: 7, Pos: 117, Val: "A"},
			{Typ: token.ASSIGN, Line: 7, Pos: 119, Val: "="},
			{Typ: token.INT, Line: 7, Pos: 121, Val: "1"},
			{Typ: token.AT, Line: 7, Pos: 123, Val: "@"},
			{Typ: token.IDENT, Line: 7, Pos: 124, Val: "ptle"},
			{Typ: token.RPAREN, Line: 8, Pos: 130, Val: ")"},
			{Typ: token.COLON, Line: 8, Pos: 131, Val: ":"},
			{Typ: token.IDENT, Line: 8, Pos: 133, Val: "Two"},
			{Typ: token.IDENT, Line: 9, Pos: 138, Val: "thr"},
			{Typ: token.COLON, Line: 9, Pos: 141, Val: ":"},
			{Typ: token.IDENT, Line: 9, Pos: 143, Val: "Thr"},
			{Typ: token.ASSIGN, Line: 9, Pos: 147, Val: "="},
			{Typ: token.INT, Line: 9, Pos: 149, Val: "3"},
			{Typ: token.AT, Line: 9, Pos: 151, Val: "@"},
			{Typ: token.IDENT, Line: 9, Pos: 152, Val: "ptle"},
			{Typ: token.AT, Line: 9, Pos: 157, Val: "@"},
			{Typ: token.IDENT, Line: 9, Pos: 158, Val: "ptle"},
			{Typ: token.RBRACE, Line: 10, Pos: 163, Val: "}"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexUnion(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`union Pizza = Triangle | Circle`), 0)
		expectItems(subT, l, []Item{
			{Typ: token.UNION, Line: 1, Pos: 0, Val: "union"},
			{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Pizza"},
			{Typ: token.ASSIGN, Line: 1, Pos: 12, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 14, Val: "Triangle"},
			{Typ: token.IDENT, Line: 1, Pos: 25, Val: "Circle"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		l := Lex("simple", []byte(`union Pizza @ham @pineapple = Triangle | Circle`), 0)
		expectItems(subT, l, []Item{
			{Typ: token.UNION, Line: 1, Pos: 0, Val: "union"},
			{Typ: token.IDENT, Line: 1, Pos: 6, Val: "Pizza"},
			{Typ: token.AT, Line: 1, Pos: 12, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 13, Val: "ham"},
			{Typ: token.AT, Line: 1, Pos: 17, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 18, Val: "pineapple"},
			{Typ: token.ASSIGN, Line: 1, Pos: 28, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 30, Val: "Triangle"},
			{Typ: token.IDENT, Line: 1, Pos: 41, Val: "Circle"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexDirective(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`directive @skip on FIELD | FIELD_DEFINITION`), 0)

		expectItems(subT, l, []Item{
			{Typ: token.DIRECTIVE, Line: 1, Pos: 0, Val: "directive"},
			{Typ: token.AT, Line: 1, Pos: 10, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 11, Val: "skip"},
			{Typ: token.ON, Line: 1, Pos: 16, Val: "on"},
			{Typ: token.IDENT, Line: 1, Pos: 19, Val: "FIELD"},
			{Typ: token.IDENT, Line: 1, Pos: 27, Val: "FIELD_DEFINITION"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withArgs", func(subT *testing.T) {
		l := Lex("simple", []byte(`directive @skip(if: Boolean, else: Boolean = false) on FIELD | FIELD_DEFINITION`), 0)

		expectItems(subT, l, []Item{
			{Typ: token.DIRECTIVE, Line: 1, Pos: 0, Val: "directive"},
			{Typ: token.AT, Line: 1, Pos: 10, Val: "@"},
			{Typ: token.IDENT, Line: 1, Pos: 11, Val: "skip"},
			{Typ: token.LPAREN, Line: 1, Pos: 15, Val: "("},
			{Typ: token.IDENT, Line: 1, Pos: 16, Val: "if"},
			{Typ: token.COLON, Line: 1, Pos: 18, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 20, Val: "Boolean"},
			{Typ: token.IDENT, Line: 1, Pos: 29, Val: "else"},
			{Typ: token.COLON, Line: 1, Pos: 33, Val: ":"},
			{Typ: token.IDENT, Line: 1, Pos: 35, Val: "Boolean"},
			{Typ: token.ASSIGN, Line: 1, Pos: 43, Val: "="},
			{Typ: token.IDENT, Line: 1, Pos: 45, Val: "false"},
			{Typ: token.RPAREN, Line: 1, Pos: 50, Val: ")"},
			{Typ: token.ON, Line: 1, Pos: 52, Val: "on"},
			{Typ: token.IDENT, Line: 1, Pos: 55, Val: "FIELD"},
			{Typ: token.IDENT, Line: 1, Pos: 63, Val: "FIELD_DEFINITION"},
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
