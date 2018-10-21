package lexer

import (
	"gqlc/sl/token"
	"testing"
)

func TestLexImports(t *testing.T) {
	t.Run("single", func(subT *testing.T) {
		subT.Run("perfect", func(triT *testing.T) {
			l := Lex("perfect", `import "hello"`)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.STRING, pos: 7, line: 1, val: `"hello"`},
			)
			expectEOF(triT, l)
		})

		subT.Run("spaces", func(triT *testing.T) {
			l := Lex("spaces", `import 		    	  	 "hello"`)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.STRING, pos: 18, line: 1, val: `"hello"`},
			)
			expectEOF(triT, l)
		})

		subT.Run("noParenOrQuote", func(triT *testing.T) {
			l := Lex("noParenOrQuote", `import hello"`)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.ERR, pos: 7, line: 1, val: `missing ( or " to begin import statement`})

		})
	})

	t.Run("multiple", func(subT *testing.T) {
		subT.Run("singleLine", func(triT *testing.T) {
			l := Lex("singleLine", `import ("hello")`)
			expectItems(triT, l,
				Item{typ: token.IMPORT, line: 1, val: "import"},
				Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
				Item{typ: token.STRING, line: 1, pos: 8, val: `"hello"`},
				Item{typ: token.RPAREN, line: 1, pos: 15, val: ")"},
			)
			expectEOF(triT, l)
		})

		subT.Run("singleLineWComma", func(triT *testing.T) {
			l := Lex("singleLineWComma", `import ( "a", "b", "c" )`)
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
			l := Lex("singleLineWCommaNoEnd", `import ( "a", "b", "c" `)
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
				`import (
					"a"
					"b"
					"c"
				)`,
			)
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
				`import (
					"a"
					"b",
					"c"
				)`,
			)
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
				`import (
					"a",
					"b"
					"c",
				)`,
			)
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
		l := Lex("simple", `scalar URI`)
		expectItems(subT, l,
			Item{typ: token.SCALAR, line: 1, val: "scalar"},
			Item{typ: token.IDENT, line: 1, pos: 7, val: "URI"},
		)
		expectEOF(subT, l)
	})

	t.Run("withDirectives", func(subT *testing.T) {
		l := Lex("withDirectives", `scalar URI @gotype @jstype() @darttype(if: Boolean)`)
		expectItems(subT, l,
			Item{typ: token.SCALAR, line: 1, val: "scalar"},
			Item{typ: token.IDENT, line: 1, pos: 7, val: "URI"},
			Item{typ: token.AT, line: 1, pos: 11, val: "@"},
			Item{typ: token.IDENT, line: 1, pos: 12, val: "gotype"},
			Item{typ: token.AT, line: 1, pos: 19, val: "@"},
			Item{typ: token.IDENT, line: 1, pos: 20, val: "jstype"},
			Item{typ: token.LPAREN, line: 1, pos: 26, val: "("},
			Item{typ: token.RPAREN, line: 1, pos: 27, val: ")"},
			Item{typ: token.AT, line: 1, pos: 29, val: "@"},
			Item{typ: token.IDENT, line: 1, pos: 30, val: "darttype"},
			Item{typ: token.LPAREN, line: 1, pos: 38, val: "("},
			Item{typ: token.IDENT, line: 1, pos: 39, val: "if"},
			Item{typ: token.COLON, line: 1, pos: 41, val: ":"},
			Item{typ: token.IDENT, line: 1, pos: 43, val: "Boolean"},
			Item{typ: token.RPAREN, line: 1, pos: 50, val: ")"},
		)
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
			input: `$a`,
		}

		go func() {
			l.scanValue()
			close(l.items)
		}()

		expectItems(subT, l,
			Item{typ: token.VAR, line: 1, pos: 0, val: "$"},
			Item{typ: token.IDENT, line: 1, pos: 1, val: "a"},
		)
	})

	t.Run("int", func(subT *testing.T) {
		l := &lxr{
			line:  1,
			items: make(chan Item),
			input: `12354654684013246813216513213254686210`,
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
				input: `123.45`,
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
				input: `123e45`,
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
				input: `123.45e6`,
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
