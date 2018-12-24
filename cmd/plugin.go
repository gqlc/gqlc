package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gqlc/compiler"
	"github.com/gqlc/graphql/ast"
	"os/exec"
	"strings"
)

type pluginGenerator struct {
	Name string
}

func (g *pluginGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	optS := strings.Split(opts, ":")

	// Create plugin request
	req := &compiler.PluginRequest{
		Docs:      []*ast.Document{doc},
		Options:   optS[0], // TODO: JSON encode
		OutputDir: optS[1],
	}

	// Encode request
	var in, out bytes.Buffer
	err := json.NewEncoder(&in).Encode(req)
	if err != nil {
		return err
	}

	// Create Command
	c := exec.Command(pluginPrefix + g.Name)
	c.Stdin = &in
	c.Stderr = &out

	// Run command
	err = c.Run()
	if err != nil {
		return err
	}

	b := out.Bytes()
	if len(b) == 0 {
		return nil
	}

	return errors.New(string(b))
}

func (g *pluginGenerator) GenerateAll(ctx context.Context, doc []*ast.Document, opts string) error {
	return nil
}
