package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"go.uber.org/zap"
)

var introQuery = `query {
	__schema {
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
	if err != nil {
		return nil, err
	}
	return newConverter(resp.Body)
}

type inputValue struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultValue string `json:"defaultValue"`
	Type         *typ   `json:"type"`
}

type field struct {
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	Args              []*inputValue `json:"args"`
	Type              *typ          `json:"type"`
	IsDeprecated      bool          `json:"isDeprecated"`
	DeprecationReason string        `json:"deprecationReason"`
}

type enum struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason"`
}

type typ struct {
	Kind          string        `json:"kind"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	OfType        *typ          `json:"ofType"`
	Fields        []*field      `json:"fields"`
	Interfaces    []*typ        `json:"interfaces"`
	PossibleTypes []*typ        `json:"possibleTypes"`
	EnumValues    []*enum       `json:"enumValues"`
	InputFields   []*inputValue `json:"inputFields"`
}

type directive struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Locations    []string      `json:"locations"`
	IsRepeatable bool          `json:"isRepeatable"`
	Args         []*inputValue `json:"args"`
}

type decodeTyp uint8

const (
	decodeDirs decodeTyp = iota
	decodeTypes
)

// converter converts the GraphQL introspection response to the GraphQL IDL
type converter struct {
	src   *json.Decoder
	close func() error

	// buffer idl in case it doesn't fit in p
	buf      bytes.Buffer
	decoding decodeTyp
}

func newConverter(rc io.ReadCloser) (*converter, error) {
	c := &converter{
		src:   json.NewDecoder(rc),
		close: rc.Close,
	}

	terr := c.init()
	return c, terr
}

func (c *converter) init() error {
	c.src.Token()
	tok, terr := c.src.Token()
	if terr != nil {
		return terr
	}

	fieldName := tok.(string)
	if fieldName != "data" {
		return fmt.Errorf("unexpected field in json: %s", fieldName)
	}
	c.src.Token()

	tok, terr = c.src.Token()
	if terr != nil {
		return terr
	}

	fieldName = tok.(string)
	if fieldName != "__schema" {
		return fmt.Errorf("expected field: \"__schema\", but got: %s", fieldName)
	}
	c.src.Token()

	tok, terr = c.src.Token()
	if terr != nil {
		return terr
	}

	fieldName = tok.(string)
	switch fieldName {
	case "directives":
		c.decoding = decodeDirs
	case "types":
		c.decoding = decodeTypes
	}
	tok, _ = c.src.Token()
	return nil
}

func (c *converter) Read(p []byte) (n int, err error) {
	if !c.src.More() {
		t, err := c.src.Token()
		if err != nil {
			return 0, err
		}

		if delim, ok := t.(json.Delim); !ok || delim != ']' {
			return 0, fmt.Errorf("expected array closing")
		}

		t, err = c.src.Token()
		if err != nil {
			return 0, err
		}
		_, ok := t.(json.Delim)
		if ok {
			return 0, io.EOF
		}

		v, ok := t.(string)
		if !ok {
			return 0, fmt.Errorf("unexpected token: %v", t)
		}
		c.src.Token()

		switch v {
		case "directives":
			c.decoding = decodeDirs
		case "types":
			c.decoding = decodeTypes
		}

		return c.Read(p)
	}

	switch c.decoding {
	case decodeDirs:
		d := directive{}
		err = c.src.Decode(&d)
		if err != nil {
			break
		}

		if d.Description != "" {
			c.buf.Write([]byte("\""))
			c.buf.WriteString(d.Description)
			c.buf.Write([]byte("\"\n"))
		}

		c.buf.Write([]byte("@"))
		c.buf.WriteString(d.Name)

		if len(d.Args) > 0 {
			c.buf.Write([]byte("("))
			for _, a := range d.Args {
				writeArg(&c.buf, a)
			}
			c.buf.Write([]byte(")"))
		}

		if d.IsRepeatable {
			c.buf.Write([]byte(" repeatable"))
		}

		c.buf.Write([]byte(" on "))

		l := len(d.Locations) - 1
		for i, loc := range d.Locations {
			c.buf.WriteString(loc)
			if i != l {
				c.buf.Write([]byte(" | "))
			}
		}
	case decodeTypes:
		t := typ{}
		err = c.src.Decode(&t)
		if err != nil {
			break
		}

		if t.Description != "" {
			c.buf.Write([]byte("\""))
			c.buf.WriteString(t.Description)
			c.buf.Write([]byte("\"\n"))
		}

		writeTyp(&c.buf, t)
	}

	return c.buf.Read(p)
}

func writeArg(b *bytes.Buffer, a *inputValue) {
	if a.Description != "" {
		b.WriteString(a.Description)
		b.Write([]byte("\n  "))
	}

	b.WriteString(a.Name)
	b.Write([]byte(": "))
	writeTypSig(b, a.Type)

	if a.DefaultValue != "" {
		b.Write([]byte(" = "))
		b.WriteString(a.DefaultValue)
	}
}

const (
	SCALAR       = "SCALAR"
	OBJECT       = "OBJECT"
	INTERFACE    = "INTERFACE"
	UNION        = "UNION"
	ENUM         = "ENUM"
	INPUT_OBJECT = "INPUT_OBJECT"
	LIST         = "LIST"
	NON_NULL     = "NON_NULL"
)

func writeTyp(b *bytes.Buffer, t typ) {
	switch t.Kind {
	case SCALAR:
		b.Write([]byte("scalar "))
		b.WriteString(t.Name)
	case OBJECT:
		b.Write([]byte("type "))
		b.WriteString(t.Name)

		if len(t.Interfaces) > 0 {
			b.Write([]byte(" implements "))
			l := len(t.Interfaces) - 1
			for i, it := range t.Interfaces {
				b.WriteString(it.Name)
				if i != l {
					b.Write([]byte(" & "))
				}
			}
		}
		b.Write([]byte(" {\n  "))

		l := len(t.Fields) - 1
		for i, f := range t.Fields {
			writeField(b, f)
			if i != l {
				b.Write([]byte("\n  "))
			}
		}
		b.Write([]byte("\n}"))
	case INTERFACE:
		b.Write([]byte("interface "))
		b.WriteString(t.Name)
		b.Write([]byte(" {\n  "))
		l := len(t.Fields) - 1
		for i, f := range t.Fields {
			writeField(b, f)
			if i != l {
				b.Write([]byte("\n  "))
			}
		}
		b.Write([]byte("\n}"))
	case UNION:
		b.Write([]byte("union "))
		b.WriteString(t.Name)
		b.Write([]byte(" = "))

		l := len(t.PossibleTypes) - 1
		for i, m := range t.PossibleTypes {
			b.WriteString(m.Name)
			if i != l {
				b.Write([]byte(" | "))
			}
		}
	case ENUM:
	case INPUT_OBJECT:
	case LIST:
	case NON_NULL:
	}
}

func writeField(b *bytes.Buffer, f *field) {
	if f.Description != "" {
		b.WriteString(f.Description)
		b.Write([]byte("\n  "))
	}
	b.WriteString(f.Name)

	if len(f.Args) > 0 {
		l := len(f.Args) - 1
		b.Write([]byte("("))
		for i, a := range f.Args {
			writeArg(b, a)
			if i != l {
				b.Write([]byte("\n    "))
			}
		}
		b.Write([]byte(")"))
	}
	b.Write([]byte(": "))

	writeTypSig(b, f.Type)
}

func writeTypSig(b *bytes.Buffer, t *typ) {
	if t.OfType == nil {
		b.WriteString(t.Name)
		return
	}

	if t.OfType.Kind == "NON_NULL" {
		b.WriteString(t.Name)
		b.Write([]byte("!"))
		return
	}

	b.Write([]byte("["))
	b.WriteString(t.Name)
	b.Write([]byte("]"))
}

func (c *converter) Close() error {
	return c.close()
}
