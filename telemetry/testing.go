package telemetry

import (
	"context"
	"testing"
)

// TestingInit initializes telemetry for testing so that it is automatically
// shutdown after the test completes.
// If ctx is nil, context.Background() is used.
func TestingInit(t *testing.T, ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	t.Cleanup(func() {
		if err := Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown telemetry: %v", err)
		}
	})
}
