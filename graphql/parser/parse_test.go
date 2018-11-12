package parser

import (
	"strings"
	"testing"
)

func TestParseDocument(t *testing.T) {
	testCases := []struct {
		Name string // test name
		Doc  string // test GraphQL document
		Err  error  // Expected error; or empty
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
			Doc: `schema @one @two {
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
			Err: nil, // TODO: Put the expected error here
		},
		{
			Name: "scalar",
			Doc:  "scalar Test @one @two",
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
		{
			Name: "directive",
			Doc:  `directive @test(one: One = 1 @one, two: Two = 2 @one @two) on SCHEMA | FIELD_DEFINITION`,
		},
		{
			Name: "extend", // TODO: Decide if this needs to be tested?
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			src := strings.NewReader(testCase.Doc)
			_, err := ParseDocument(src, ParseComments)
			if err != testCase.Err {
				subT.Error(err)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	// TODO
}
