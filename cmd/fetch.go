package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/zaba505/gws"
	"go.uber.org/zap"
)

var introQuery = `query IntrospectionQuery {
      __schema {
        types {
          ...FullType
        }
        directives {
          name
          description
          locations
          args {
            ...InputValue
          }
        }
      }
    }

    fragment FullType on __Type {
      kind
      name
      description
      fields(includeDeprecated: true) {
        name
        description
        args {
          ...InputValue
        }
        type {
          ...TypeRef
        }
        isDeprecated
        deprecationReason
      }
      inputFields {
        ...InputValue
      }
      interfaces {
        ...TypeRef
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
      possibleTypes {
        ...TypeRef
      }
    }

    fragment InputValue on __InputValue {
      name
      description
      type { ...TypeRef }
      defaultValue
    }

    fragment TypeRef on __Type {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                  ofType {
                    kind
                    name
                  }
                }
              }
            }
          }
        }
      }
    }`

type gqlReq struct {
	Query string `json:"query"`
}

var query bytes.Buffer

func init() {
	err := json.NewEncoder(&query).Encode(gqlReq{Query: introQuery})
	if err != nil {
		zap.L().Error("unexpected error when encoding introspection query", zap.Error(err))
		return
	}
}

type fetchClient struct {
	*http.Client
}

func fetch(client *fetchClient, url *url.URL) (io.ReadCloser, error) {
	if strings.HasPrefix(url.Scheme, "ws") || filepath.Base(url.Path) == "graphql" {
		zap.L().Info("fetching types via introspection", zap.String("endpoint", url.String()))
		return client.introspect(url)
	}

	zap.L().Info("fetching remote file", zap.String("name", url.String()))
	resp, err := client.Get(url.String())
	return resp.Body, err
}

type noopCloser struct {
	io.Reader
}

func (noopCloser) Close() error { return nil }

func (c *fetchClient) introspect(endpoint *url.URL) (io.ReadCloser, error) {
	var resp *gws.Response

	switch endpoint.Scheme {
	case "http", "https":
		r, err := c.Post(endpoint.String(), "application/json", &query)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		resp = new(gws.Response)
		err = json.Unmarshal(b, resp)
		if err != nil {
			return nil, err
		}
	case "ws", "wss":
		conn, err := gws.Dial(context.TODO(), endpoint.String())
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		wc := gws.NewClient(conn)

		resp, err = wc.Query(context.TODO(), &gws.Request{Query: introQuery})
		if err != nil {
			return nil, err
		}
	default:
		// TODO
		return nil, nil
	}
	// TODO: Check resp.Errors

	return newConverter(noopCloser{bytes.NewReader(resp.Data)})
}
