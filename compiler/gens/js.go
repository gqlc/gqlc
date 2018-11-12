package gens

import (
	"context"
	"gqlc/graphql/ast"
)

// JsGenerator generates Javascript code for a GraphQL schema.
type JsGenerator struct{}

func (gen JsGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

func (gen JsGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
