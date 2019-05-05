package plugin

import (
	"bytes"
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"os/exec"
)

// Generator is a plugin generator.
type Generator struct {
	Name   string
	Prefix string
}

// Generate executes a plugin given the GraphQL Document.
func (g *Generator) Generate(ctx context.Context, doc *ast.Document, opts string) (err error) {
	defer func() {
		if err != nil {
			err = compiler.Error{
				GenName: g.Name,
				DocName: doc.Name,
				Msg:     err.Error(),
			}
		}
	}()

	// Marshall doc
	b, perr := proto.Marshal(&PluginRequest{
		FileToGenerate: []string{doc.Name},
		Parameter:      opts,
		Documents:      []*ast.Document{doc},
	})
	if perr != nil {
		err = perr
		return
	}

	// Create plugin command
	out := new(bytes.Buffer)
	pluginName := g.Prefix + g.Name
	cmd := exec.Command(pluginName)
	cmd.Stdin = bytes.NewReader(b)
	cmd.Stdout = out

	// Exec plugin
	err = cmd.Run()
	if err != nil {
		return
	}

	// Unmarshall response
	var resp PluginResponse
	err = proto.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return
	}

	// Check response
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	// Write plugin files
	gCtx := compiler.Context(ctx)
	for _, f := range resp.File {
		w, ferr := gCtx.Open(f.Name)
		if ferr != nil {
			err = ferr
			return
		}

		_, err = w.Write([]byte(f.Content))
		if err != nil {
			return
		}
	}
	return
}
