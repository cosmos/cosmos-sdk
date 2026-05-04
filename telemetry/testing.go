package telemetry

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// TestingMain should be used in tests where you want to run telemetry and need clean shutdown
// behavior at the end of the test, for instance to collect benchmark metrics.
// If ctx is nil, context.Background() is used.
// Example:
//
//	func TestMain(m *testing.M) {
//	    telemetry.TestingMain(m, nil)
//	}
func TestingMain(m *testing.M, ctx context.Context) {
	code := m.Run()
	if ctx == nil {
		ctx = context.Background()
	}
	if err := Shutdown(ctx); err != nil {
		fmt.Printf("failed to shutdown telemetry after test completion: %v\n", err)
	}
	os.Exit(code)
}
