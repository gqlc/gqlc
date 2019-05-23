// Package plugin contains a Generator for running external plugins as Generators.
package plugin

import (
	"bytes"
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/gqlc/compiler"
	gqlc "github.com/gqlc/gqlc/cmd/plugin/proto"
	"github.com/gqlc/graphql/ast"
	"os/exec"
)

// Generator executes an external plugin as a generator.
// The name of the plugin is given by the generators Prefix and Name fields.
//
type Generator struct {
	Name   string
	Prefix string

	*exec.Cmd
}

// Generate executes a plugin given the GraphQL Document.
func (g *Generator) Generate(ctx context.Context, doc *ast.Document, opts string) (err error) {
	defer func() {
		if err != nil {
			err = compiler.GeneratorError{
				GenName: g.Name,
				DocName: doc.Name,
				Msg:     err.Error(),
			}
		}
	}()

	// Marshall doc
	b, perr := proto.Marshal(&gqlc.PluginRequest{
		FileToGenerate: []string{doc.Name},
		Parameter:      opts,
		Documents:      []*ast.Document{doc},
	})
	if perr != nil {
		err = perr
		return
	}

	// Create plugin command
	pluginName := g.Prefix + g.Name
	if g.Cmd == nil {
		g.Cmd = exec.CommandContext(ctx, pluginName)
	}
	out := new(bytes.Buffer)
	g.Stdin = bytes.NewReader(b)
	g.Stdout = out

	// Exec plugin
	err = g.Run()
	g.Cmd = nil
	if err != nil {
		return
	}

	// Unmarshall response
	var resp gqlc.PluginResponse
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
