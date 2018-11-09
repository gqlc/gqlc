package gens

import (
	"context"
	"gqlc/graphql/file"
)

// DartGenerator generates Dart code for a GraphQL schema.
type DartGenerator struct{}

func (gen DartGenerator) Generate(ctx context.Context, file *file.Descriptor, opts string) error {
	return nil
}

func (gen DartGenerator) GenerateAll(ctx context.Context, files []*file.Descriptor, opts string) error {
	return nil
}
