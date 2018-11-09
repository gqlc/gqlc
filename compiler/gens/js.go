package gens

import (
	"context"
	"gqlc/sl/file"
)

// JsGenerator generates Javascript code for a GraphQL schema.
type JsGenerator struct{}

func (gen JsGenerator) Generate(ctx context.Context, file *file.Descriptor, opts string) error {
	return nil
}

func (gen JsGenerator) GenerateAll(ctx context.Context, files []*file.Descriptor, opts string) error {
	return nil
}
