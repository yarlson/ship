package testctx

import (
	"context"
	"testing"
	"time"
)

const defaultTimeout = 2 * time.Minute

// New returns a test-scoped context with a deadline so blocking operations fail fast.
func New(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	t.Cleanup(cancel)

	return ctx
}

// Background returns a timeout-bounded context for setup outside an individual test.
func Background() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}
