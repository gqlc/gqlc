package gens

import (
	"context"
	"gqlc/graphql/file"
)

// GoGenerator generates Go code for a GraphQL schema.
type GoGenerator struct{}

func (gen GoGenerator) Generate(ctx context.Context, file *file.Descriptor, opts string) error {
	return nil
}

func (gen GoGenerator) GenerateAll(ctx context.Context, files []*file.Descriptor, opts string) error {
	return nil
}
