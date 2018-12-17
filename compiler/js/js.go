// Package js contains a Javascript generator for GraphQL Documents.
package js

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// Generator generates Javascript code for a GraphQL schema.
type Generator struct{}

// Generate generates Javascript code for all schemas found within the given document.
func (gen *Generator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

// GenerateAll generates Javascript code for all schemas found within all the given documents.
func (gen *Generator) GenerateAll(ctx context.Context, docs []*ast.Document, opts string) error {
	return nil
}
