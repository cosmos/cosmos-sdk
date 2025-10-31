package telemetry

import (
	"context"
	"testing"
)

// TestingInit initializes telemetry for testing.
// If ctx is nil, context.Background() is used.
// If logger is nil, a new test logger is created.
func TestingInit(t *testing.T, ctx context.Context) {
	t.Cleanup(func() {
		if err := Shutdown(ctx); err != nil {
			t.Fatalf("failed to shutdown telemetry: %v", err)
		}
	})
}
