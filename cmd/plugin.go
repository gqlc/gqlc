package cmd

import (
	"context"
	"github.com/gqlc/graphql/ast"
)

type pluginGenerator struct {
	Name   string
	Prefix string
}

func (g *pluginGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}
