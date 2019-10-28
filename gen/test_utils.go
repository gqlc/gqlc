package gen

import "io"

// TestCtx is a noop closer, which wraps an io.Writer
// and only meant to be used for tests.
//
type TestCtx struct {
	io.Writer
}

// Open returns the underlying io.Writer.
func (ctx TestCtx) Open(filename string) (io.WriteCloser, error) { return ctx, nil }

// Close always returns nil.
func (ctx TestCtx) Close() error { return nil }
