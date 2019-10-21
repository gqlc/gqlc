package gen

import "io"

// TestCtx
type TestCtx struct {
	io.Writer
}

// Open
func (ctx TestCtx) Open(filename string) (io.WriteCloser, error) { return ctx, nil }

// Close
func (ctx TestCtx) Close() error { return nil }
