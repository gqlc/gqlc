package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

var testGqlData = []byte(`
{
  "data": {
    "__schema": {
      "description": null,
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
		w.Write(testGqlData)
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

	// After fetching it should convert to GraphQL IDL
	// Hence, equal testGqlFile
	if !bytes.Equal(b, testGqlFile) {
		t.Fail()
		return
	}
}

func TestConverter(t *testing.T) {

}
