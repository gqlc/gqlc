package compiler

import (
	"context"
	"gqlc/sl/file"
)

// CodeGenerator provides a simple API for creating a code generator for
// any language desired. It also represents a `gqlc` plugin, where the input
// schema is JSON value sent to the plugin STDIN.
type CodeGenerator interface {
	// Generate should handle multiple schemas in a single file.
	Generate(ctx context.Context, file *file.Descriptor, opts string) error

	// GenerateAll should handle multiple schemas.
	GenerateAll(ctx context.Context, files []*file.Descriptor, opts string) error
}
