package gens

import (
	"context"
	"gqlc/graphql/ast"
)

// DartGenerator generates Dart code for a GraphQL schema.
type DartGenerator struct{}

func (gen DartGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

func (gen DartGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
