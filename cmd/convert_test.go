package cmd

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestConverter(t *testing.T) {
	testCases := []struct {
		Name string
		JSON string
		IDL  []byte
	}{
		{
			Name: "SCALAR",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "SCALAR",
			          "name": "Time",
			          "description": null,
			          "fields": null,
			          "interfaces": null,
			          "possibleTypes": null,
			          "enumValues": null,
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte("scalar Time\n"),
		},
		{
			Name: "OBJECT",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "OBJECT",
			          "name": "Test",
			          "description": null,
			          "fields": [
									{
										"name": "a",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "b",
												"description": null,
												"isDeprecated": false,
												"deprecationReason": null,
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "String",
											"ofType": null
										}
									},
									{
										"name": "b",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "LIST"
											}
										}
									},
									{
										"name": "c",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "d",
												"description": null,
												"defaultValue": "0",
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "NON_NULL"
											}
										}
									}
								],
			          "interfaces": [
									{
										"name": "A"
									},
									{
										"name": "B"
									}
								],
			          "possibleTypes": null,
			          "enumValues": null,
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte(`type Test implements A & B {
  a(b: Int): String
  b: [Int]
  c(d: Int = 0): Int!
}
`),
		},
		{
			Name: "OBJECT with Descriptions",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "OBJECT",
			          "name": "Test",
			          "description": null,
			          "fields": [
									{
										"name": "a",
										"description": "a is a field.",
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "b",
												"description": "b is a arg.",
												"isDeprecated": false,
												"deprecationReason": null,
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "String",
											"ofType": null
										}
									},
									{
										"name": "b",
										"description": "b is a field.",
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "LIST"
											}
										}
									},
									{
										"name": "c",
										"description": "c is a field.",
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "d",
												"description": "d is a arg.",
												"defaultValue": "0",
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "NON_NULL"
											}
										}
									}
								],
			          "interfaces": [
									{
										"name": "A"
									},
									{
										"name": "B"
									}
								],
			          "possibleTypes": null,
			          "enumValues": null,
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte(`type Test implements A & B {
  "a is a field." a("b is a arg." b: Int): String
  "b is a field." b: [Int]
  "c is a field." c("d is a arg." d: Int = 0): Int!
}
`),
		},
		{
			Name: "INTERFACE",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "INTERFACE",
			          "name": "Test",
			          "description": null,
								"fields": [
									{
										"name": "a",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "b",
												"description": null,
												"isDeprecated": false,
												"deprecationReason": null,
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "String",
											"ofType": null
										}
									},
									{
										"name": "b",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "LIST"
											}
										}
									},
									{
										"name": "c",
										"description": null,
										"isDeprecated": false,
										"deprecationReason": null,
										"args": [
											{
												"name": "d",
												"description": null,
												"defaultValue": "0",
												"type": {
													"kind": "SCALAR",
													"name": "Int",
													"ofType": null
												}
											}
										],
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "NON_NULL"
											}
										}
									}
								],
			          "interfaces": null,
			          "possibleTypes": null,
			          "enumValues": null,
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte(`interface Test {
  a(b: Int): String
  b: [Int]
  c(d: Int = 0): Int!
}
`),
		},
		{
			Name: "UNION",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "UNION",
			          "name": "Test",
			          "description": null,
								"fields": [],
			          "interfaces": null,
			          "possibleTypes": [
									{
										"kind": "OBJECT",
										"name": "A"
									},
									{
										"kind": "OBJECT",
										"name": "B"
									},
									{
										"kind": "OBJECT",
										"name": "C"
									}
								],
			          "enumValues": null,
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte("union Test = A | B | C\n"),
		},
		{
			Name: "ENUM",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "ENUM",
			          "name": "Test",
			          "description": null,
								"fields": [],
			          "interfaces": null,
								"possibleTypes": null,
			          "enumValues": [
									{
										"name": "A",
										"description": null,
										"isDeprecated": null,
										"deprecationReason": null
									},
									{
										"name": "B",
										"description": null,
										"isDeprecated": null,
										"deprecationReason": null
									},
									{
										"name": "C",
										"description": null,
										"isDeprecated": null,
										"deprecationReason": null
									}
								],
			          "inputFields": null,
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte(`enum Test {
  A
  B
  C
}
`),
		},
		{
			Name: "INPUT",
			JSON: `
			{
			    "__schema": {
			      "directives": [],
			      "types": [
			        {
			          "kind": "INPUT_OBJECT",
			          "name": "Test",
			          "description": null,
								"fields": [],
			          "interfaces": null,
			          "possibleTypes": null,
			          "enumValues": null,
			          "inputFields": [
									{
										"name": "a",
										"description": null,
										"defaultValue": null,
										"type": {
											"kind": "SCALAR",
											"name": "String",
											"ofType": null
										}
									},
									{
										"name": "b",
										"description": null,
										"defaultValue": null,
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": {
												"kind": "LIST"
											}
										}
									},
									{
										"name": "d",
										"description": null,
										"defaultValue": "0",
										"type": {
											"kind": "SCALAR",
											"name": "Int",
											"ofType": null
										}
									}
								],
			          "ofType": null
			        }
			      ]
			    }
			}
			`,
			IDL: []byte(`input Test {
  a: String
  b: [Int]
  d: Int = 0
}
`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			rc := noopCloser{strings.NewReader(testCase.JSON)}
			c, err := newConverter(rc)
			if err != nil {
				t.Errorf("unexpected error when initing converter: %s", err)
				return
			}

			b, err := ioutil.ReadAll(c)
			if err != nil {
				subT.Errorf("unexpected error when converting: %s", err)
				return
			}

			if !bytes.Equal(b, testCase.IDL) {
				t.Logf("\nexpected: %s\ngot: %s", string(testCase.IDL), string(b))
				t.Fail()
				return
			}
		})
	}
}
