package api

import "context"

// DoUntilCtxExpired runs the given function until the context is expired or
// the function exits.
// This forces context to be honored.
func DoUntilCtxExpired(ctx context.Context, f func()) error {
	done := make(chan struct{})
	go func() {
		defer close(done)

		f()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
