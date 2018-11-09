// Package gens contains generator implementations for various languages
package gens

import (
	"context"
	"gqlc/sl/file"
)

// DocGenerator generates Documentation for a GraphQL schema.
type DocGenerator struct{}

// Generate generates documentation for all schemas found within the given file.
func (gen DocGenerator) Generate(ctx context.Context, file *file.Descriptor, opts string) error {
	return nil
}

// GenerateAll generates documentation for all schemas found within all the given files.
func (gen DocGenerator) GenerateAll(ctx context.Context, files []*file.Descriptor, opts string) error {
	return nil
}
