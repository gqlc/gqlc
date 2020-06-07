package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testGqlFile = []byte(`scalar Time`)

func TestFetch_RemoteFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(testGqlFile)
	}))
	defer srv.Close()

	r, err := fetch(http.DefaultClient, "http://"+srv.Listener.Addr().String())
	if err != nil {
		t.Errorf("unexpected error when fetching file: %s", err)
		return
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("unexpected error when reading response: %s", err)
		return
	}

	if !bytes.Equal(b, testGqlFile) {
		t.Fail()
		return
	}
}

var testRespData = []byte(`
{
  "data": {
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
}
`)

func TestFetch_FromService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(testRespData)
	}))
	defer srv.Close()

	r, err := fetch(http.DefaultClient, fmt.Sprintf("http://%s/graphql", srv.Listener.Addr().String()))
	if err != nil {
		t.Errorf("unexpected error when fetching file: %s", err)
		return
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("unexpected error when reading response: %s", err)
		return
	}

	// After fetching it should convert the response to the GraphQL IDL.
	// Hence, equal testGqlFile
	if !bytes.Equal(b, testGqlFile) {
		t.Fail()
		return
	}
}

type noopCloser struct {
	io.Reader
}

func (noopCloser) Close() error { return nil }

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
			  "data": {
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
			}
			`,
			IDL: []byte("scalar Time"),
		},
		{
			Name: "OBJECT",
			JSON: `
			{
			  "data": {
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
			}
			`,
			IDL: []byte(`type Test implements A & B {
  a(b: Int): String
  b: [Int]
  c(d: Int = 0): Int!
}`),
		},
		{
			Name: "INTERFACE",
			JSON: `
			{
			  "data": {
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
			}
			`,
			IDL: []byte(`interface Test {
  a(b: Int): String
  b: [Int]
  c(d: Int = 0): Int!
}`),
		},
		{
			Name: "UNION",
			JSON: `
			{
			  "data": {
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
			}
			`,
			IDL: []byte(`union Test = A | B | C`),
		},
		{
			Name: "ENUM",
			JSON: `
			{
			  "data": {
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
			}
			`,
			IDL: []byte(`enum Test {
  A
  B
  C
}`),
		},
		{
			Name: "INPUT",
			JSON: `
			{
			  "data": {
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
			}
			`,
			IDL: []byte(`input Test {
  a: String
  b: [Int]
  d: Int = 0
}`),
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
