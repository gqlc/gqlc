package gen

import (
	"bytes"
	"io"
	"testing"
)

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

// CompareBytes is a testing utility for comparing generator outputs.
func CompareBytes(t *testing.T, ex, out []byte) {
	if bytes.EqualFold(out, ex) {
		return
	}

	line := 1
	for i, b := range out {
		if b == '\n' {
			line++
		}

		if ex[i] != b {
			t.Fatalf("expected: %s, but got: %s, %d:%d", string(ex[i]), string(b), i, line)
		}
	}
}
