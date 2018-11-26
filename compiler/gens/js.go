package gens

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// JsGenerator generates Javascript code for a GraphQL schema.
type JsGenerator struct{}

// Generate generates Javascript code for all schemas found within the given document.
func (gen JsGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

// GenerateAll generates Javascript code for all schemas found within all the given documents.
func (gen JsGenerator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
