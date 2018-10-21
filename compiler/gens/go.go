package gens

import "context"

type GoGenerator struct{}

func (gen GoGenerator) Generate(ctx context.Context) error {
	return nil
}

func (gen GoGenerator) GenerateAll(ctx context.Context) error {
	return nil
}
