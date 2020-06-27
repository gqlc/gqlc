package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/zaba505/gws"
)

var testGqlFile = []byte("scalar Time\n")

func TestFetch_RemoteFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write(testGqlFile)
	}))
	defer srv.Close()

	endpoint, _ := url.Parse("http://" + srv.Listener.Addr().String())

	r, err := fetch(&fetchClient{Client: http.DefaultClient}, endpoint, nil)
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
`)

func TestFetch_FromService(t *testing.T) {
	wh := gws.NewHandler(gws.HandlerFunc(func(s *gws.Stream, req *gws.Request) error {
		s.Send(context.TODO(), &gws.Response{Data: []byte(testRespData)})
		return s.Close()
	}))

	m := http.NewServeMux()
	m.Handle("/", wh)
	m.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		b, _ := json.Marshal(&gws.Response{Data: []byte(testRespData)})
		w.Write(b)
	})

	testCases := []struct {
		Name   string
		Scheme string
		Path   string
	}{
		{
			Name:   "Over HTTP",
			Scheme: "http",
			Path:   "graphql",
		},
		{
			Name:   "Over Websocket",
			Scheme: "ws",
		},
	}

	srv := httptest.NewServer(m)
	defer srv.Close()

	testClient := &fetchClient{
		Client: http.DefaultClient,
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(subT *testing.T) {
			endpoint, _ := url.Parse(fmt.Sprintf("%s://%s/%s", testCase.Scheme, srv.Listener.Addr().String(), testCase.Path))

			r, err := fetch(testClient, endpoint, nil)
			if err != nil {
				subT.Errorf("unexpected error when fetching file: %s", err)
				return
			}
			defer r.Close()

			b, err := ioutil.ReadAll(r)
			if err != nil {
				subT.Errorf("unexpected error when reading response: %s", err)
				return
			}

			// After fetching it should convert the response to the GraphQL IDL.
			// Hence, equal testGqlFile
			if !bytes.Equal(b, testGqlFile) {
				subT.Fail()
				return
			}
		})
	}
}

func TestFetch_WithHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Hello") == "" {
			t.Log("expected header: Hello")
			t.Fail()
		}

		b, _ := json.Marshal(&gws.Response{Data: []byte(testRespData)})
		w.Write(b)
	}))
	defer srv.Close()

	testClient := &fetchClient{
		Client: http.DefaultClient,
	}

	endpoint, _ := url.Parse(fmt.Sprintf("http://%s/graphql", srv.Listener.Addr().String()))

	r, err := fetch(testClient, endpoint, http.Header{"hello": []string{"world"}})
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
