// Package doc contains a Documentation generator for GraphQL Documents.
package doc

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// Generator generates Documentation for a GraphQL schema.
type Generator struct{}

// Generate generates documentation for all schemas found within the given document.
func (gen *Generator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	// Create one large markdown source
	// Pass markdown source through html renderer if option is passed
	return nil
}

// GenerateAll generates documentation for all schemas found within all the given documents.
func (gen *Generator) GenerateAll(ctx context.Context, doc []*ast.Document, opts string) error {
	return nil
}
