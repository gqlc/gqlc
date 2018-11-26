// Package gens contains generator implementations for various languages
package gens

import (
	"context"
	"github.com/Zaba505/gqlc/graphql/ast"
)

// DocGenerator generates Documentation for a GraphQL schema.
type DocGenerator struct{}

// Generate generates documentation for all schemas found within the given document.
func (gen DocGenerator) Generate(ctx context.Context, doc *ast.Document, opts string) error {
	return nil
}

// GenerateAll generates documentation for all schemas found within all the given documents.
func (gen DocGenerator) GenerateAll(ctx context.Context, doc []*ast.Document, opts string) error {
	return nil
}
