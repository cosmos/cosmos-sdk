package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
)

// TestingInit initializes telemetry for testing.
// If ctx is nil, context.Background() is used.
// If logger is nil, a new test logger is created.
func TestingInit(t *testing.T, ctx context.Context, logger log.Logger) *Metrics {
	t.Helper()
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = log.NewTestLogger(t)
	}

	// configure metrics and tracing for testing
	telemetryCfg := Config{
		Enabled:     true,
		ServiceName: "cosmos-sdk-test",
	}
	telemetryCfgJson, ok := os.LookupEnv("COSMOS_TELEMETRY")
	if ok && telemetryCfgJson != "" {
		err := json.Unmarshal([]byte(telemetryCfgJson), &telemetryCfg)
		require.NoError(t, err, "failed to parse telemetry config", telemetryCfgJson)
	}

	t.Logf("Configuring telemetry with: %+v", telemetryCfg)
	metrics, err := New(telemetryCfg, WithLogger(logger))
	require.NoError(t, err, "failed to initialize telemetry")
	err = metrics.Start(ctx)
	require.NoError(t, err, "failed to start telemetry")
	t.Cleanup(func() {
		require.NoError(t, metrics.Shutdown(ctx), "failed to shutdown telemetry")
	})
	return metrics
}
