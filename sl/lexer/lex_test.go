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
				{typ: token.IMPORT, line: 1, val: "import"},
				{typ: token.STRING, pos: 7, line: 1, val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("spaces", func(triT *testing.T) {
			l := Lex("spaces", []byte(`import 		    	  	 "hello"`), 0)
			expectItems(triT, l, []Item{
				{typ: token.IMPORT, line: 1, val: "import"},
				{typ: token.STRING, pos: 18, line: 1, val: `"hello"`},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("noParenOrQuote", func(triT *testing.T) {
			l := Lex("noParenOrQuote", []byte(`import hello"`), 0)
			expectItems(triT, l, []Item{
				{typ: token.IMPORT, line: 1, val: "import"},
				{typ: token.ERR, pos: 7, line: 1, val: `missing ( or " to begin import statement`},
			}...)

		})
	})

	t.Run("multiple", func(subT *testing.T) {
		subT.Run("singleLine", func(triT *testing.T) {
			l := Lex("singleLine", []byte(`import ("hello")`), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 1, pos: 8, val: `"hello"`},
				Item{typ: token.RPAREN, line: 1, pos: 15, val: ")"},
			)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWComma", func(triT *testing.T) {
			l := Lex("singleLineWComma", []byte(`import ( "a", "b", "c" )`), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 1, pos: 9, val: `"a"`},
				Item{typ: token.STRING, line: 1, pos: 14, val: `"b"`},
				Item{typ: token.STRING, line: 1, pos: 19, val: `"c"`},
				Item{typ: token.RPAREN, line: 1, pos: 23, val: ")"},
			)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWCommaNoEnd", func(triT *testing.T) {
			l := Lex("singleLineWCommaNoEnd", []byte(`import ( "a", "b", "c" `), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 1, pos: 9, val: `"a"`},
				Item{typ: token.STRING, line: 1, pos: 14, val: `"b"`},
				Item{typ: token.STRING, line: 1, pos: 19, val: `"c"`},
				Item{typ: token.ERR, line: 1, pos: 23, val: `invalid list seperator: -1`},
			)
		})

		subT.Run("multiLinesWNewLine", func(triT *testing.T) {
			l := Lex("multLinesWNewLine",
				[]byte(`import (
					"a"
					"b"
					"c"
				)`), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
				Item{typ: token.STRING, line: 3, pos: 23, val: `"b"`},
				Item{typ: token.STRING, line: 4, pos: 32, val: `"c"`},
				Item{typ: token.RPAREN, line: 5, pos: 40, val: ")"},
			)
			expectEOF(triT, l)
		})

		subT.Run("multiLinesWDiffSep1", func(triT *testing.T) {
			l := Lex("multLinesWDiffSep1",
				[]byte(`import (
					"a"
					"b",
					"c"
				)`), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
				Item{typ: token.STRING, line: 3, pos: 23, val: `"b"`},
				Item{typ: token.ERR, line: 3, pos: 26, val: `list seperator must remain the same throughout the list`},
			)
		})

		subT.Run("multiLinesWDiffSep2", func(triT *testing.T) {
			l := Lex("multLinesWDiffSep2",
				[]byte(`import (
					"a",
					"b"
					"c",
				)`), 0)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
				Item{typ: token.STRING, line: 3, pos: 24, val: `"b"`},
				Item{typ: token.ERR, line: 3, pos: 27, val: `list seperator must remain the same throughout the list`},
			)
		})

	})
}

func TestLexScalar(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`scalar URI`), 0)
		expectItems(subT, l, []Item{
			{typ: token.SCALAR, line: 1, val: "scalar"},
			{typ: token.IDENT, line: 1, pos: 7, val: "URI"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		l := Lex("withDirectives", []byte(`scalar URI @gotype @jstype() @darttype(if: Boolean)`), 0)
		expectItems(subT, l, []Item{
			{typ: token.SCALAR, line: 1, val: "scalar"},
			{typ: token.IDENT, line: 1, pos: 7, val: "URI"},
			{typ: token.AT, line: 1, pos: 11, val: "@"},
			{typ: token.IDENT, line: 1, pos: 12, val: "gotype"},
			{typ: token.AT, line: 1, pos: 19, val: "@"},
			{typ: token.IDENT, line: 1, pos: 20, val: "jstype"},
			{typ: token.LPAREN, line: 1, pos: 26, val: "("},
			{typ: token.RPAREN, line: 1, pos: 27, val: ")"},
			{typ: token.AT, line: 1, pos: 29, val: "@"},
			{typ: token.IDENT, line: 1, pos: 30, val: "darttype"},
			{typ: token.LPAREN, line: 1, pos: 38, val: "("},
			{typ: token.IDENT, line: 1, pos: 39, val: "if"},
			{typ: token.COLON, line: 1, pos: 41, val: ":"},
			{typ: token.IDENT, line: 1, pos: 43, val: "Boolean"},
			{typ: token.RPAREN, line: 1, pos: 50, val: ")"},
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
			{typ: token.VAR, line: 1, pos: 0, val: "$"},
			{typ: token.IDENT, line: 1, pos: 1, val: "a"},
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
			Item{typ: token.INT, line: 1, pos: 0, val: "12354654684013246813216513213254686210"},
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
				Item{typ: token.FLOAT, line: 1, pos: 0, val: "123.45"},
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
				Item{typ: token.FLOAT, line: 1, pos: 0, val: "123e45"},
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
				Item{typ: token.FLOAT, line: 1, pos: 0, val: "123.45e6"},
			)
		})
	})
}

func TestLexObject(t *testing.T) {
	t.Run("withImpls", func(subT *testing.T) {
		subT.Run("perfect", func(triT *testing.T) {
			l := Lex("perfect", []byte(`type Rect implements One & Two & Three`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.IDENT, line: 1, pos: 27, val: "Two"},
				{typ: token.IDENT, line: 1, pos: 33, val: "Three"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("invalidSeperator", func(triT *testing.T) {
			l := Lex("first", []byte(`type Rect implements One , Two & Three`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.ERR, line: 1, pos: 25, val: "invalid list seperator: 44"},
			}...)

			l = Lex("later", []byte(`type Rect implements One & Two , Three`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.IDENT, line: 1, pos: 27, val: "Two"},
				{typ: token.ERR, line: 1, pos: 31, val: "invalid list seperator: 44"},
			}...)
		})
	})

	t.Run("withDirectives", func(subT *testing.T) {
		subT.Run("endsWithBrace", func(triT *testing.T) {
			l := Lex("endsWithBrace", []byte(`type Rect @green @blue {}`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.AT, line: 1, pos: 10, val: "@"},
				{typ: token.IDENT, line: 1, pos: 11, val: "green"},
				{typ: token.AT, line: 1, pos: 17, val: "@"},
				{typ: token.IDENT, line: 1, pos: 18, val: "blue"},
				{typ: token.LBRACE, line: 1, pos: 23, val: "{"},
				{typ: token.RBRACE, line: 1, pos: 24, val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			l := Lex("endsWithNewline", []byte(`type Rect @green @blue
`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.AT, line: 1, pos: 10, val: "@"},
				{typ: token.IDENT, line: 1, pos: 11, val: "green"},
				{typ: token.AT, line: 1, pos: 17, val: "@"},
				{typ: token.IDENT, line: 1, pos: 18, val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			l := Lex("endsWithEOF", []byte(`type Rect @green @blue`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.AT, line: 1, pos: 10, val: "@"},
				{typ: token.IDENT, line: 1, pos: 11, val: "green"},
				{typ: token.AT, line: 1, pos: 17, val: "@"},
				{typ: token.IDENT, line: 1, pos: 18, val: "blue"},
			}...)
			expectEOF(triT, l)
		})
	})

	t.Run("withImpls&Directives", func(subT *testing.T) {
		subT.Run("endsWithBrace", func(triT *testing.T) {
			l := Lex("endsWithBrace", []byte(`type Rect implements One & Two & Three @green @blue {}`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.IDENT, line: 1, pos: 27, val: "Two"},
				{typ: token.IDENT, line: 1, pos: 33, val: "Three"},
				{typ: token.AT, line: 1, pos: 39, val: "@"},
				{typ: token.IDENT, line: 1, pos: 40, val: "green"},
				{typ: token.AT, line: 1, pos: 46, val: "@"},
				{typ: token.IDENT, line: 1, pos: 47, val: "blue"},
				{typ: token.LBRACE, line: 1, pos: 52, val: "{"},
				{typ: token.RBRACE, line: 1, pos: 53, val: "}"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithNewline", func(triT *testing.T) {
			l := Lex("endsWithNewline", []byte(`type Rect implements One & Two & Three @green @blue
`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.IDENT, line: 1, pos: 27, val: "Two"},
				{typ: token.IDENT, line: 1, pos: 33, val: "Three"},
				{typ: token.AT, line: 1, pos: 39, val: "@"},
				{typ: token.IDENT, line: 1, pos: 40, val: "green"},
				{typ: token.AT, line: 1, pos: 46, val: "@"},
				{typ: token.IDENT, line: 1, pos: 47, val: "blue"},
			}...)
			expectEOF(triT, l)
		})

		subT.Run("endsWithEOF", func(triT *testing.T) {
			l := Lex("endsWithEOF", []byte(`type Rect implements One & Two & Three @green @blue`), 0)
			expectItems(triT, l, []Item{
				{typ: token.TYPE, line: 1, pos: 0, val: "type"},
				{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
				{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
				{typ: token.IDENT, line: 1, pos: 21, val: "One"},
				{typ: token.IDENT, line: 1, pos: 27, val: "Two"},
				{typ: token.IDENT, line: 1, pos: 33, val: "Three"},
				{typ: token.AT, line: 1, pos: 39, val: "@"},
				{typ: token.IDENT, line: 1, pos: 40, val: "green"},
				{typ: token.AT, line: 1, pos: 46, val: "@"},
				{typ: token.IDENT, line: 1, pos: 47, val: "blue"},
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
					{typ: token.TYPE, line: 1, pos: 0, val: "type"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.IDENT, line: 2, pos: 13, val: "one"},
					{typ: token.COLON, line: 2, pos: 16, val: ":"},
					{typ: token.IDENT, line: 2, pos: 18, val: "One"},
					{typ: token.IDENT, line: 3, pos: 23, val: "two"},
					{typ: token.COLON, line: 3, pos: 26, val: ":"},
					{typ: token.IDENT, line: 3, pos: 28, val: "Two"},
					{typ: token.RBRACE, line: 4, pos: 32, val: "}"},
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
					{typ: token.TYPE, line: 1, pos: 0, val: "type"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.DESCRIPTION, line: 2, pos: 13, val: `"one descr"`},
					{typ: token.IDENT, line: 2, pos: 25, val: "one"},
					{typ: token.COLON, line: 2, pos: 28, val: ":"},
					{typ: token.IDENT, line: 2, pos: 30, val: "One"},
					{typ: token.DESCRIPTION, line: 5, pos: 35, val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
					{typ: token.IDENT, line: 6, pos: 56, val: "two"},
					{typ: token.COLON, line: 6, pos: 59, val: ":"},
					{typ: token.IDENT, line: 6, pos: 61, val: "Two"},
					{typ: token.RBRACE, line: 7, pos: 65, val: "}"},
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
					{typ: token.TYPE, line: 1, pos: 0, val: "type"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.IDENT, line: 2, pos: 13, val: "one"},
					{typ: token.LPAREN, line: 2, pos: 16, val: "("},
					{typ: token.IDENT, line: 2, pos: 17, val: "a"},
					{typ: token.COLON, line: 2, pos: 18, val: ":"},
					{typ: token.IDENT, line: 2, pos: 20, val: "A"},
					{typ: token.IDENT, line: 2, pos: 23, val: "b"},
					{typ: token.COLON, line: 2, pos: 24, val: ":"},
					{typ: token.IDENT, line: 2, pos: 26, val: "B"},
					{typ: token.RPAREN, line: 2, pos: 27, val: ")"},
					{typ: token.COLON, line: 2, pos: 28, val: ":"},
					{typ: token.IDENT, line: 2, pos: 30, val: "One"},
					{typ: token.IDENT, line: 3, pos: 35, val: "two"},
					{typ: token.LPAREN, line: 3, pos: 38, val: "("},
					{typ: token.DESCRIPTION, line: 4, pos: 41, val: `"a descr"`},
					{typ: token.IDENT, line: 4, pos: 51, val: "a"},
					{typ: token.COLON, line: 4, pos: 52, val: ":"},
					{typ: token.IDENT, line: 4, pos: 54, val: "A"},
					{typ: token.DESCRIPTION, line: 7, pos: 57, val: "\"\"\"\n\tb descr\n\t\"\"\""},
					{typ: token.IDENT, line: 8, pos: 76, val: "b"},
					{typ: token.COLON, line: 8, pos: 77, val: ":"},
					{typ: token.IDENT, line: 8, pos: 79, val: "B"},
					{typ: token.RPAREN, line: 9, pos: 81, val: ")"},
					{typ: token.COLON, line: 9, pos: 82, val: ":"},
					{typ: token.IDENT, line: 9, pos: 84, val: "Two"},
					{typ: token.RBRACE, line: 10, pos: 88, val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				l := Lex("withDirectives", []byte(`type Rect {
	one: One @green @blue
	two: Two @blue
}`), 0)
				expectItems(qt, l, []Item{
					{typ: token.TYPE, line: 1, pos: 0, val: "type"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.IDENT, line: 2, pos: 13, val: "one"},
					{typ: token.COLON, line: 2, pos: 16, val: ":"},
					{typ: token.IDENT, line: 2, pos: 18, val: "One"},
					{typ: token.AT, line: 2, pos: 22, val: "@"},
					{typ: token.IDENT, line: 2, pos: 23, val: "green"},
					{typ: token.AT, line: 2, pos: 29, val: "@"},
					{typ: token.IDENT, line: 2, pos: 30, val: "blue"},
					{typ: token.IDENT, line: 3, pos: 36, val: "two"},
					{typ: token.COLON, line: 3, pos: 39, val: ":"},
					{typ: token.IDENT, line: 3, pos: 41, val: "Two"},
					{typ: token.AT, line: 3, pos: 45, val: "@"},
					{typ: token.IDENT, line: 3, pos: 46, val: "blue"},
					{typ: token.RBRACE, line: 4, pos: 51, val: "}"},
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
					{typ: token.ENUM, line: 1, pos: 0, val: "enum"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.IDENT, line: 2, pos: 13, val: "LEFT"},
					{typ: token.IDENT, line: 3, pos: 19, val: "UP"},
					{typ: token.IDENT, line: 4, pos: 23, val: "RIGHT"},
					{typ: token.IDENT, line: 5, pos: 30, val: "DOWN"},
					{typ: token.RBRACE, line: 6, pos: 35, val: "}"},
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
					{typ: token.ENUM, line: 1, pos: 0, val: "enum"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.DESCRIPTION, line: 2, pos: 13, val: `"left descr"`},
					{typ: token.IDENT, line: 2, pos: 26, val: "LEFT"},
					{typ: token.DESCRIPTION, line: 3, pos: 32, val: `"up descr"`},
					{typ: token.IDENT, line: 3, pos: 43, val: "UP"},
					{typ: token.DESCRIPTION, line: 6, pos: 47, val: "\"\"\"\n\tright descr\n\t\"\"\""},
					{typ: token.IDENT, line: 7, pos: 70, val: "RIGHT"},
					{typ: token.DESCRIPTION, line: 8, pos: 77, val: `"down descr"`},
					{typ: token.IDENT, line: 9, pos: 91, val: "DOWN"},
					{typ: token.RBRACE, line: 10, pos: 96, val: "}"},
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
					{typ: token.ENUM, line: 1, pos: 0, val: "enum"},
					{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 10, val: "{"},
					{typ: token.IDENT, line: 2, pos: 13, val: "LEFT"},
					{typ: token.AT, line: 2, pos: 18, val: "@"},
					{typ: token.IDENT, line: 2, pos: 19, val: "green"},
					{typ: token.AT, line: 2, pos: 25, val: "@"},
					{typ: token.IDENT, line: 2, pos: 26, val: "blue"},
					{typ: token.IDENT, line: 3, pos: 32, val: "UP"},
					{typ: token.AT, line: 3, pos: 35, val: "@"},
					{typ: token.IDENT, line: 3, pos: 36, val: "red"},
					{typ: token.IDENT, line: 4, pos: 41, val: "RIGHT"},
					{typ: token.IDENT, line: 5, pos: 48, val: "DOWN"},
					{typ: token.AT, line: 5, pos: 53, val: "@"},
					{typ: token.IDENT, line: 5, pos: 54, val: "red"},
					{typ: token.AT, line: 5, pos: 58, val: "@"},
					{typ: token.IDENT, line: 5, pos: 59, val: "green"},
					{typ: token.AT, line: 5, pos: 65, val: "@"},
					{typ: token.IDENT, line: 5, pos: 66, val: "blue"},
					{typ: token.RBRACE, line: 6, pos: 71, val: "}"},
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
					{typ: token.INPUT, line: 1, pos: 0, val: "input"},
					{typ: token.IDENT, line: 1, pos: 6, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 11, val: "{"},
					{typ: token.IDENT, line: 2, pos: 14, val: "one"},
					{typ: token.COLON, line: 2, pos: 17, val: ":"},
					{typ: token.IDENT, line: 2, pos: 19, val: "One"},
					{typ: token.IDENT, line: 3, pos: 24, val: "two"},
					{typ: token.COLON, line: 3, pos: 27, val: ":"},
					{typ: token.IDENT, line: 3, pos: 29, val: "Two"},
					{typ: token.RBRACE, line: 4, pos: 33, val: "}"},
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
					{typ: token.INPUT, line: 1, pos: 0, val: "input"},
					{typ: token.IDENT, line: 1, pos: 6, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 11, val: "{"},
					{typ: token.DESCRIPTION, line: 2, pos: 14, val: `"one descr"`},
					{typ: token.IDENT, line: 2, pos: 26, val: "one"},
					{typ: token.COLON, line: 2, pos: 29, val: ":"},
					{typ: token.IDENT, line: 2, pos: 31, val: "One"},
					{typ: token.DESCRIPTION, line: 5, pos: 36, val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
					{typ: token.IDENT, line: 6, pos: 57, val: "two"},
					{typ: token.COLON, line: 6, pos: 60, val: ":"},
					{typ: token.IDENT, line: 6, pos: 62, val: "Two"},
					{typ: token.RBRACE, line: 7, pos: 66, val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDefVal", func(qt *testing.T) {
				l := Lex("withDefVal", []byte(`input Rect {
	one: One = 123
	two: Two = "abc"
}`), 0)
				expectItems(qt, l, []Item{
					{typ: token.INPUT, line: 1, pos: 0, val: "input"},
					{typ: token.IDENT, line: 1, pos: 6, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 11, val: "{"},
					{typ: token.IDENT, line: 2, pos: 14, val: "one"},
					{typ: token.COLON, line: 2, pos: 17, val: ":"},
					{typ: token.IDENT, line: 2, pos: 19, val: "One"},
					{typ: token.ASSIGN, line: 2, pos: 23, val: "="},
					{typ: token.INT, line: 2, pos: 25, val: "123"},
					{typ: token.IDENT, line: 3, pos: 30, val: "two"},
					{typ: token.COLON, line: 3, pos: 33, val: ":"},
					{typ: token.IDENT, line: 3, pos: 35, val: "Two"},
					{typ: token.ASSIGN, line: 3, pos: 39, val: "="},
					{typ: token.STRING, line: 3, pos: 41, val: `"abc"`},
					{typ: token.RBRACE, line: 4, pos: 47, val: "}"},
				}...)
				expectEOF(qt, l)
			})

			triT.Run("withDirectives", func(qt *testing.T) {
				l := Lex("withDirectives", []byte(`input Rect {
	one: One @green @blue
	two: Two @blue
}`), 0)
				expectItems(qt, l, []Item{
					{typ: token.INPUT, line: 1, pos: 0, val: "input"},
					{typ: token.IDENT, line: 1, pos: 6, val: "Rect"},
					{typ: token.LBRACE, line: 1, pos: 11, val: "{"},
					{typ: token.IDENT, line: 2, pos: 14, val: "one"},
					{typ: token.COLON, line: 2, pos: 17, val: ":"},
					{typ: token.IDENT, line: 2, pos: 19, val: "One"},
					{typ: token.AT, line: 2, pos: 23, val: "@"},
					{typ: token.IDENT, line: 2, pos: 24, val: "green"},
					{typ: token.AT, line: 2, pos: 30, val: "@"},
					{typ: token.IDENT, line: 2, pos: 31, val: "blue"},
					{typ: token.IDENT, line: 3, pos: 37, val: "two"},
					{typ: token.COLON, line: 3, pos: 40, val: ":"},
					{typ: token.IDENT, line: 3, pos: 42, val: "Two"},
					{typ: token.AT, line: 3, pos: 46, val: "@"},
					{typ: token.IDENT, line: 3, pos: 47, val: "blue"},
					{typ: token.RBRACE, line: 4, pos: 52, val: "}"},
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
			{typ: token.TYPE, line: 1, pos: 0, val: "type"},
			{typ: token.IDENT, line: 1, pos: 5, val: "Rect"},
			{typ: token.IMPLEMENTS, line: 1, pos: 10, val: "implements"},
			{typ: token.IDENT, line: 1, pos: 21, val: "Shape"},
			{typ: token.IDENT, line: 1, pos: 29, val: "Obj"},
			{typ: token.AT, line: 1, pos: 33, val: "@"},
			{typ: token.IDENT, line: 1, pos: 34, val: "green"},
			{typ: token.AT, line: 1, pos: 40, val: "@"},
			{typ: token.IDENT, line: 1, pos: 41, val: "blue"},
			{typ: token.LBRACE, line: 1, pos: 46, val: "{"},
			{typ: token.DESCRIPTION, line: 2, pos: 49, val: `"one descr"`},
			{typ: token.IDENT, line: 2, pos: 61, val: "one"},
			{typ: token.COLON, line: 2, pos: 64, val: ":"},
			{typ: token.IDENT, line: 2, pos: 66, val: "One"},
			{typ: token.AT, line: 2, pos: 70, val: "@"},
			{typ: token.IDENT, line: 2, pos: 71, val: "one"},
			{typ: token.DESCRIPTION, line: 5, pos: 76, val: "\"\"\"\n\ttwo descr\n\t\"\"\""},
			{typ: token.IDENT, line: 6, pos: 97, val: "two"},
			{typ: token.LPAREN, line: 6, pos: 100, val: "("},
			{typ: token.DESCRIPTION, line: 7, pos: 104, val: `"a descr"`},
			{typ: token.IDENT, line: 7, pos: 114, val: "a"},
			{typ: token.COLON, line: 7, pos: 115, val: ":"},
			{typ: token.IDENT, line: 7, pos: 117, val: "A"},
			{typ: token.ASSIGN, line: 7, pos: 119, val: "="},
			{typ: token.INT, line: 7, pos: 121, val: "1"},
			{typ: token.AT, line: 7, pos: 123, val: "@"},
			{typ: token.IDENT, line: 7, pos: 124, val: "ptle"},
			{typ: token.RPAREN, line: 8, pos: 130, val: ")"},
			{typ: token.COLON, line: 8, pos: 131, val: ":"},
			{typ: token.IDENT, line: 8, pos: 133, val: "Two"},
			{typ: token.IDENT, line: 9, pos: 138, val: "thr"},
			{typ: token.COLON, line: 9, pos: 141, val: ":"},
			{typ: token.IDENT, line: 9, pos: 143, val: "Thr"},
			{typ: token.ASSIGN, line: 9, pos: 147, val: "="},
			{typ: token.INT, line: 9, pos: 149, val: "3"},
			{typ: token.AT, line: 9, pos: 151, val: "@"},
			{typ: token.IDENT, line: 9, pos: 152, val: "ptle"},
			{typ: token.AT, line: 9, pos: 157, val: "@"},
			{typ: token.IDENT, line: 9, pos: 158, val: "ptle"},
			{typ: token.RBRACE, line: 10, pos: 163, val: "}"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexUnion(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`union Pizza = Triangle | Circle`), 0)
		expectItems(subT, l, []Item{
			{typ: token.UNION, line: 1, pos: 0, val: "union"},
			{typ: token.IDENT, line: 1, pos: 6, val: "Pizza"},
			{typ: token.ASSIGN, line: 1, pos: 12, val: "="},
			{typ: token.IDENT, line: 1, pos: 14, val: "Triangle"},
			{typ: token.IDENT, line: 1, pos: 25, val: "Circle"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		l := Lex("simple", []byte(`union Pizza @ham @pineapple = Triangle | Circle`), 0)
		expectItems(subT, l, []Item{
			{typ: token.UNION, line: 1, pos: 0, val: "union"},
			{typ: token.IDENT, line: 1, pos: 6, val: "Pizza"},
			{typ: token.AT, line: 1, pos: 12, val: "@"},
			{typ: token.IDENT, line: 1, pos: 13, val: "ham"},
			{typ: token.AT, line: 1, pos: 17, val: "@"},
			{typ: token.IDENT, line: 1, pos: 18, val: "pineapple"},
			{typ: token.ASSIGN, line: 1, pos: 28, val: "="},
			{typ: token.IDENT, line: 1, pos: 30, val: "Triangle"},
			{typ: token.IDENT, line: 1, pos: 41, val: "Circle"},
		}...)
		expectEOF(subT, l)
	})
}

func TestLexDirective(t *testing.T) {
	t.Run("simple", func(subT *testing.T) {
		l := Lex("simple", []byte(`directive @skip on FIELD | FIELD_DEFINITION`), 0)

		expectItems(subT, l, []Item{
			{typ: token.DIRECTIVE, line: 1, pos: 0, val: "directive"},
			{typ: token.AT, line: 1, pos: 10, val: "@"},
			{typ: token.IDENT, line: 1, pos: 11, val: "skip"},
			{typ: token.ON, line: 1, pos: 16, val: "on"},
			{typ: token.IDENT, line: 1, pos: 19, val: "FIELD"},
			{typ: token.IDENT, line: 1, pos: 27, val: "FIELD_DEFINITION"},
		}...)
		expectEOF(subT, l)
	})

	t.Run("withArgs", func(subT *testing.T) {
		l := Lex("simple", []byte(`directive @skip(if: Boolean, else: Boolean = false) on FIELD | FIELD_DEFINITION`), 0)

		expectItems(subT, l, []Item{
			{typ: token.DIRECTIVE, line: 1, pos: 0, val: "directive"},
			{typ: token.AT, line: 1, pos: 10, val: "@"},
			{typ: token.IDENT, line: 1, pos: 11, val: "skip"},
			{typ: token.LPAREN, line: 1, pos: 15, val: "("},
			{typ: token.IDENT, line: 1, pos: 16, val: "if"},
			{typ: token.COLON, line: 1, pos: 18, val: ":"},
			{typ: token.IDENT, line: 1, pos: 20, val: "Boolean"},
			{typ: token.IDENT, line: 1, pos: 29, val: "else"},
			{typ: token.COLON, line: 1, pos: 33, val: ":"},
			{typ: token.IDENT, line: 1, pos: 35, val: "Boolean"},
			{typ: token.ASSIGN, line: 1, pos: 43, val: "="},
			{typ: token.IDENT, line: 1, pos: 45, val: "false"},
			{typ: token.RPAREN, line: 1, pos: 50, val: ")"},
			{typ: token.ON, line: 1, pos: 52, val: "on"},
			{typ: token.IDENT, line: 1, pos: 55, val: "FIELD"},
			{typ: token.IDENT, line: 1, pos: 63, val: "FIELD_DEFINITION"},
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
	if i.typ != token.EOF {
		t.Fatalf("expected eof but instead received: %#v", i)
	}
}
