// Package gen contains and utils for working with generators.
package gen

//go:generate mockgen -write_package_comment=false -package=gen -destination=./mock.go github.com/gqlc/gqlc/gen Generator

import (
	"context"
	"fmt"
	"io"

	"github.com/gqlc/graphql/ast"
)

// Generator provides a simple API for creating a code generator for
// any language desired.
//
type Generator interface {
	// Generate handles converting a GraphQL Document to scaffolded source code.
	Generate(ctx context.Context, doc *ast.Document, opts map[string]interface{}) error
}

// GeneratorContext represents the directory to which
// the Generator is to write to.
//
type GeneratorContext interface {
	// Open opens a file in the GeneratorContext (i.e. directory).
	Open(filename string) (io.WriteCloser, error)
}

type genCtx string

var genCtxKey = genCtx("genCtx")

// WithContext returns a prepared context.Context
// with the given GeneratorContext.
//
func WithContext(ctx context.Context, gCtx GeneratorContext) context.Context {
	return context.WithValue(ctx, genCtxKey, gCtx)
}

// Context returns the generator context.
func Context(ctx context.Context) GeneratorContext {
	return ctx.Value(genCtxKey).(GeneratorContext)
}

// GeneratorError represents an error from a generator.
type GeneratorError struct {
	// DocName is the document being worked on when error was encountered.
	DocName string

	// GenName is the generator name which encountered a problem.
	GenName string

	// Msg is any message the generator wants to provide back to the caller.
	Msg string
}

func (e GeneratorError) Error() string {
	return fmt.Sprintf("compiler: generator error occurred in %s:%s %s", e.GenName, e.DocName, e.Msg)
}
