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
