package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

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

	maxRetries uint8
}

var defaultClient = &fetchClient{
	Client: &http.Client{
		Timeout: 1 * time.Minute,
	},
	maxRetries: 5,
}

func fetch(client *fetchClient, url *url.URL, headers http.Header) (io.ReadCloser, error) {
	if strings.HasPrefix(url.Scheme, "ws") || filepath.Base(url.Path) == "graphql" {
		zap.L().Info("fetching types via introspection", zap.String("endpoint", url.String()), zap.Any("headers", headers))
		return client.introspect(url, headers)
	}

	req, _ := http.NewRequest(http.MethodGet, url.String(), nil)
	req.Header = headers

	zap.L().Info("fetching remote file", zap.String("name", url.String()), zap.Any("headers", headers))
	resp, err := client.Do(req)
	return resp.Body, err
}

type noopCloser struct {
	io.Reader
}

func (noopCloser) Close() error { return nil }

func (c *fetchClient) introspect(endpoint *url.URL, headers http.Header) (io.ReadCloser, error) {
	var resp *gws.Response

	switch endpoint.Scheme {
	case "http", "https":
		hs := make(http.Header)
		for k, v := range headers {
			for _, s := range v {
				hs.Add(k, s)
			}
		}
		hs.Set("Content-Type", "application/json")

		req, _ := http.NewRequest(http.MethodPost, endpoint.String(), &query)
		req.Header = hs

		r, err := c.Do(req)
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
		conn, err := gws.Dial(context.TODO(), endpoint.String(), gws.WithHeaders(headers))
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

func (c *fetchClient) Do(req *http.Request) (resp *http.Response, err error) {
	var b []byte
	if req.Body != nil {
		b, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}

	body := bytes.NewReader(b)

	attempt := 0
	for attempt < int(c.maxRetries) {
		body.Seek(0, 0)

		attempt++
		timeout := time.Duration(math.Exp2(float64(attempt))-1) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		r := req.WithContext(ctx)
		r.Body = &noopCloser{body}

		zap.L().Info("performing http request", zap.String("endpoint", req.URL.String()), zap.Int("attempt", attempt), zap.Duration("timeout", timeout))
		resp, err = c.Client.Do(r)

		if err == nil {
			b, err = ioutil.ReadAll(resp.Body)
			cancel()
			resp.Body = &noopCloser{bytes.NewReader(b)}
			return
		}

		cancel()
		if _, ok := err.(*url.Error); !ok {
			return
		}
	}

	return nil, context.DeadlineExceeded
}
