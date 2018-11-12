package gens

import (
	"context"
	"gqlc/graphql/ast"
)

// GoGenerator generates Go code for a GraphQL schema.
type GoGenerator struct{}

func (gen GoGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

func (gen GoGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
