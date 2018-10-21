// Package gens contains generator implementations for various languages
package gens

import "context"

// DocGenerator generates Documentation from the GraphQL schema
type DocGenerator struct{}

func (gen DocGenerator) Generate(ctx context.Context) error {
	return nil
}

func (gen DocGenerator) GenerateAll(ctx context.Context) error {
	return nil
}
