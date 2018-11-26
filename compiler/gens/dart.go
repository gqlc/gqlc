package gens

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// DartGenerator generates Dart code for a GraphQL schema.
type DartGenerator struct{}

// Generate generates Dart code for all schemas found within the given document.
func (gen DartGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

// GenerateAll generates Dart code for all schemas found within all the given documents.
func (gen DartGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
