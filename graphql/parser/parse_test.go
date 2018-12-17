package parser

import (
	"encoding/json"
	"fmt"
	"github.com/Zaba505/gqlc/graphql/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDoc(t *testing.T) {
	testCases := []struct {
		Name  string // test name
		Doc   string // test GraphQL document
		Err   error  // Expected error; or empty
		Print bool   // Helper for checking ast structure
	}{
		{
			Name: "justImports",
			Doc: `import (
	"one.gql"
	"two.gql"
)`,
		},
		{
			Name: "schema",
			Doc: `schema @one @two() @three(a: "A") {
	query: Query!
	mutation: Mutation
	subscription: Subscription
}`,
		},
		{
			Name: "invalidSchema",
			Doc: `schema {
	query: Query
	mut: Mutation
}`,
			Err: fmt.Errorf("parser: invalidSchema:3: unexpected \"mut\" in parseSchema"),
		},
		{
			Name: "scalar",
			Doc:  "scalar Test @one @two() @three(a: 1, b: 2)",
		},
		{
			Name: "object",
			Doc: `type Test implements One & Two @one @two {
				one(): One @one @two
				two(one: One): Two! @one @two
				thr(one: One = 1, two: Two): [Thr]! @one @two
				for(one: One = 1 @one @two, two: Two = 2 @one @two, thr: Thr = 3 @one @two): [For!]! @one @two
			}`,
		},
		/*
			{
				Name: "interface",
				Doc: `interface One @one @two {
					one(): One @one @two
					two(one: One): Two! @one @two
					thr(one: One = 1, two: Two): [Thr]! @one @two
					for(one: One = 1 @one @two, two: Two = 2 @one @two, thr: Thr = 3 @one @two): [For!]! @one @two
				}`,
			},
			{
				Name: "union",
				Doc:  "union Test @one @two = One | Two | Three",
			},
			{
				Name: "enum",
				Doc: `enum Test @one @two {
					"One before" ONE @one
					"""
					Two above
					"""
					TWO	@one @two
					"Three above"
					"Three before" THREE @one @two @three
				}`,
			},
			{
				Name: "input",
				Doc: `input Test @one @two {
					one: One @one
					two: Two = 2 @one @two
				}`,
			},
		*/
		{
			Name: "directive",
			Doc:  `directive @test(one: One = 1 @one, two: Two = 2 @one @two) on SCHEMA | FIELD_DEFINITION`,
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			src := strings.NewReader(testCase.Doc)
			d, err := ParseDoc(token.NewDocSet(), testCase.Name, src, 0)
			if testCase.Err != nil {
				if err.Error() == testCase.Err.Error() {
					return
				}
				subT.Error(err)
				return
			}
			if err != nil {
				subT.Errorf("unexpected error from parser: %s", err)
			}

			if !testCase.Print {
				return
			}

			if err = enc.Encode(d); err != nil {
				subT.Errorf("unexpected error while marshalling ast.Document to json: %s", err)
			}
		})
	}
}

func TestParseDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("unable to get current working directory: %s", err)
		return
	}

	dset := token.NewDocSet()
	docs, err := ParseDir(dset, filepath.Join(wd, "testdir"), func(info os.FileInfo) bool {
		return info.IsDir() && info.Name() == "skipdir"
	}, 0)

	if err != nil {
		t.Errorf("unexpected error encountered when parsing directory, 'testdir': %s", err)
		return
	}

	doc, ok := docs["test.gql"]
	if !ok {
		fmt.Println(docs)
		t.Fail()
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	if err = enc.Encode(doc); err != nil {
		t.Errorf("unexpected error while marshalling ast.Document to json: %s", err)
	}
}
