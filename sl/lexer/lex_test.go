package lexer

import (
	"gqlc/sl/token"
	"testing"
)

func TestLexImports(t *testing.T) {
	t.Run("single", func(subT *testing.T) {
		l := Lex("perfect", `import "hello"`)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.STRING, pos: 7, line: 1, val: `"hello"`})

		l = Lex("spaces", `import 		    	  	 "hello"`)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.STRING, pos: 18, line: 1, val: `"hello"`})

		l = Lex("noParenOrQuote", `import hello"`)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.ERR, pos: 7, line: 1, val: `missing ( or " to begin import statement`})
	})

	t.Run("multiple", func(subT *testing.T) {
		l := Lex("singleLine", `import ("hello")`)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 1, pos: 8, val: `"hello"`},
			Item{typ: token.RPAREN, line: 1, pos: 15, val: ")"},
		)

		l = Lex("singleLineWComma", `import ( "a", "b", "c" )`)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 1, pos: 9, val: `"a"`},
			Item{typ: token.STRING, line: 1, pos: 14, val: `"b"`},
			Item{typ: token.STRING, line: 1, pos: 19, val: `"c"`},
			Item{typ: token.RPAREN, line: 1, pos: 23, val: ")"},
		)

		l = Lex("singleLineWCommaNoEnd", `import ( "a", "b", "c" `)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 1, pos: 9, val: `"a"`},
			Item{typ: token.STRING, line: 1, pos: 14, val: `"b"`},
			Item{typ: token.STRING, line: 1, pos: 19, val: `"c"`},
			Item{typ: token.ERR, line: 1, pos: 22, val: `invalid list seperator: -1`},
		)

		l = Lex("multLinesWNewLine",
			`import (
					"a"
					"b"
					"c"
				)`,
		)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
			Item{typ: token.STRING, line: 3, pos: 23, val: `"b"`},
			Item{typ: token.STRING, line: 4, pos: 32, val: `"c"`},
			Item{typ: token.RPAREN, line: 5, pos: 40, val: ")"},
		)

		l = Lex("multLinesWDiffSep1",
			`import (
					"a"
					"b",
					"c"
				)`,
		)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
			Item{typ: token.STRING, line: 3, pos: 23, val: `"b"`},
			Item{typ: token.ERR, line: 3, pos: 26, val: `list seperator must remain the same throughout the list`},
		)

		l = Lex("multLinesWDiffSep1",
			`import (
					"a",
					"b"
					"c",
				)`,
		)
		checkItems(subT, l,
			Item{typ: token.IMPORT, line: 1, val: "import"},
			Item{typ: token.LPAREN, line: 1, pos: 7, val: "("},
			Item{typ: token.STRING, line: 2, pos: 14, val: `"a"`},
			Item{typ: token.STRING, line: 3, pos: 24, val: `"b"`},
			Item{typ: token.ERR, line: 3, pos: 27, val: `list seperator must remain the same throughout the list`},
		)

	})
}

func checkItems(t *testing.T, l Interface, items ...Item) {
	for _, item := range items {
		lItem := l.NextItem()
		if lItem != item {
			t.Errorf("expected item: %#v but instead received: %#v", item, lItem)
		}
	}
}
