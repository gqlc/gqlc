package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"

	"go.uber.org/zap"
)

var introQuery = `query {
	__schema {
		description
		types {
			kind
			name
			description
      ofType {
        kind
      }

			fields(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
				args {
					name
					description
					defaultValue
					type {
						kind
						name

            ofType {
              kind
            }
					}
				}
				type {
					kind
					name

          ofType {
            kind
          }
				}
			}

			interfaces {
				name
			}

			possibleTypes {
				kind
				name
			}

			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}

			inputFields {
				name
				description
				defaultValue
				type {
					kind
					name

          ofType {
            kind
          }
				}
			}
		}
		directives {
			name
			description
			locations
			isRepeatable
			args {
				name
				description
				defaultValue
				type {
					kind
					name

          ofType {
            kind
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

func fetch(client *http.Client, url string) (io.ReadCloser, error) {
	if filepath.Base(url) != "graphql" {
		zap.L().Info("fetching remote file", zap.String("name", url))
		resp, err := client.Get(url)
		return resp.Body, err
	}

	resp, err := client.Post(url, "application/json", &query)
	return &converter{resp.Body}, err
}

// converter converts the GraphQL introspection response to the GraphQL IDL
type converter struct {
	src io.ReadCloser
}

func (c *converter) Read(p []byte) (int, error) {
	return c.src.Read(p)
}

func (c *converter) Close() error {
	return c.src.Close()
}
