package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := log.NewNopLogger()
	enableTelemetry := func() {}

	t.Run("create server with default config", func(t *testing.T) {
		srv, err := New[transaction.Tx](nil, logger, enableTelemetry)
		require.NoError(t, err)
		require.NotNil(t, srv)
		require.Equal(t, ServerName, srv.Name())
		require.NotNil(t, srv.metrics)
	})

	t.Run("create server with custom config", func(t *testing.T) {
		cfg := map[string]interface{}{
			"telemetry": map[string]interface{}{
				"enable":  true,
				"address": "localhost:9090",
			},
		}
		srv, err := New[transaction.Tx](cfg, logger, enableTelemetry)
		require.NoError(t, err)
		require.NotNil(t, srv)
		require.Equal(t, "localhost:9090", srv.config.Address)
	})

	t.Run("fail with nil enableTelemetry", func(t *testing.T) {
		require.Panics(t, func() {
			_, _ = New[transaction.Tx](nil, logger, nil)
		})
	})
}

func TestServer_Lifecycle(t *testing.T) {
	logger := log.NewNopLogger()
	enableTelemetry := func() {}

	t.Run("start and stop enabled server", func(t *testing.T) {
		srv, err := New[transaction.Tx](nil, logger, enableTelemetry)
		require.NoError(t, err)

		ctx := context.Background()
		go func() {
			err := srv.Start(ctx)
			require.NoError(t, err)
		}()

		// Allow server to start
		time.Sleep(100 * time.Millisecond)

		err = srv.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("start disabled server", func(t *testing.T) {
		cfg := map[string]interface{}{
			"telemetry": map[string]interface{}{
				"enable": false,
			},
		}
		srv, err := New[transaction.Tx](cfg, logger, enableTelemetry)
		require.NoError(t, err)

		err = srv.Start(context.Background())
		require.NoError(t, err)
	})
}

func TestServer_MetricsHandler(t *testing.T) {
	logger := log.NewNopLogger()
	enableTelemetry := func() {}

	srv, err := New[transaction.Tx](nil, logger, enableTelemetry)
	require.NoError(t, err)

	t.Run("default metrics endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		srv.metricsHandler(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestServer_Config(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		srv := NewWithConfigOptions[transaction.Tx]()
		cfg := srv.Config().(*Config)
		require.NotNil(t, cfg)
		require.Equal(t, DefaultConfig().Address, cfg.Address)
	})

	t.Run("config with options", func(t *testing.T) {
		customAddr := "localhost:8081"
		srv := NewWithConfigOptions[transaction.Tx](func(cfg *Config) {
			cfg.Address = customAddr
		})
		cfg := srv.Config().(*Config)
		require.Equal(t, customAddr, cfg.Address)
	})
}
