// Package plugin contains a Generator for running external plugins as Generators.
package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gqlc/gqlc/gen"
	"github.com/gqlc/gqlc/plugin/pb"
	"github.com/gqlc/graphql/ast"
	"go.uber.org/zap"
)

// Generator executes an external plugin as a generator.
// The name of the plugin is given by the generators Prefix and Name fields.
//
type Generator struct {
	*exec.Cmd

	Name   string
	Prefix string

	lookOnce    sync.Once
	path        string
	lookPathErr error
	log         *zap.Logger
}

// Generate executes a plugin given the GraphQL Document.
func (g *Generator) Generate(ctx context.Context, doc *ast.Document, opts map[string]interface{}) (err error) {
	defer func() {
		if err != nil {
			err = gen.GeneratorError{
				GenName: g.Prefix + g.Name,
				DocName: doc.Name,
				Msg:     err.Error(),
			}
		}
	}()

	if g.log == nil {
		g.log = zap.L().Named(g.Name).With(zap.String("doc", doc.Name))
	}

	// Encode options to JSON
	g.log.Info("marshalling options")
	b, err := json.Marshal(opts)
	if err != nil {
		return err
	}

	// Lookup plugin only once
	g.lookOnce.Do(func() {
		pluginName := g.Prefix + g.Name
		g.path, g.lookPathErr = exec.LookPath(pluginName)
	})
	if g.lookPathErr != nil {
		err = g.lookPathErr
		return
	}

	// Marshall doc
	g.log.Info("marshalling request")
	b, perr := proto.Marshal(&pb.Request{
		FileToGenerate: []string{doc.Name},
		Parameter:      string(b),
		Documents:      []*ast.Document{doc},
	})
	if perr != nil {
		err = perr
		return
	}

	// Configure plugin command
	if g.Cmd == nil {
		g.Cmd = exec.CommandContext(ctx, g.path)
	}
	out := new(bytes.Buffer)
	g.Stdin = bytes.NewReader(b)
	g.Stdout = out

	// Exec plugin
	g.log.Info("executing plugin")
	err = g.Run()
	g.Cmd = nil
	if err != nil {
		return
	}

	// Unmarshall response
	g.log.Info("unmarshalling response")
	var resp pb.Response
	err = proto.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return
	}

	// Check response
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	// Write plugin files
	gCtx := gen.Context(ctx)
	for _, f := range resp.File {
		g.log.Info("writing content from plugin", zap.String("file", f.Name))

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
