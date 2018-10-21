package compiler

import "context"

// CodeGenerator provides a simple API for creating code generator for
// any language desired
type CodeGenerator interface {
	Generate(ctx context.Context) error
	GenerateAll(ctx context.Context) error
}
