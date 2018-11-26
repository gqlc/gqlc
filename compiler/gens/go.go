package gens

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// GoGenerator generates Go code for a GraphQL schema.
type GoGenerator struct{}

// GenerateAll generates Go code for all schemas found within the given document.
func (gen GoGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

// GenerateAll generates Go code for all schemas found within all the given documents.
func (gen GoGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
