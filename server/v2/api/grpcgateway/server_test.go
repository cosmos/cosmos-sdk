package grpcgateway_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/api/grpcgateway"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := log.NewTestLogger(t)

	t.Run("create server with default config", func(t *testing.T) {
		server, err := grpcgateway.New[transaction.Tx](logger, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, server)
		require.Equal(t, "grpc-gateway", server.Name())
	})

	t.Run("create server with custom config", func(t *testing.T) {
		cfg := map[string]any{
			"grpc-gateway": map[string]any{
				"address": "localhost:8081",
				"enable":  true,
			},
		}
		server, err := grpcgateway.New[transaction.Tx](logger, cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, server)
	})
}

func TestServerLifecycle(t *testing.T) {
	logger := log.NewTestLogger(t)

	t.Run("disabled server lifecycle", func(t *testing.T) {
		server, err := grpcgateway.New[transaction.Tx](
			logger,
			map[string]any{
				"grpc-gateway": map[string]any{
					"enable": false,
				},
			},
			nil,
		)
		require.NoError(t, err)

		ctx := context.Background()
		require.NoError(t, server.Start(ctx))
		require.NoError(t, server.Stop(ctx))
	})

	t.Run("enabled server lifecycle", func(t *testing.T) {
		server, err := grpcgateway.New[transaction.Tx](
			logger,
			map[string]any{
				"grpc-gateway": map[string]any{
					"address": "localhost:0",
					"enable":  true,
				},
			},
			nil,
		)
		require.NoError(t, err)

		ctx := context.Background()
		go func() {
			_ = server.Start(ctx)
		}()

		time.Sleep(100 * time.Millisecond)
		require.NoError(t, server.Stop(ctx))
	})
}

func TestCustomGRPCHeaderMatcher(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		expectedKey   string
		shouldInclude bool
	}{
		{
			name:          "block height header",
			header:        "x-cosmos-block-height",
			expectedKey:   "x-cosmos-block-height",
			shouldInclude: true,
		},
		{
			name:          "standard header",
			header:        "content-type",
			expectedKey:   "grpcgateway-Content-Type",
			shouldInclude: true,
		},
		{
			name:          "custom header not matched",
			header:        "x-custom-header",
			expectedKey:   "",
			shouldInclude: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, include := grpcgateway.CustomGRPCHeaderMatcher(tc.header)
			require.Equal(t, tc.shouldInclude, include)
			if tc.shouldInclude {
				require.Equal(t, tc.expectedKey, key)
			}
		})
	}
}

func TestConfigOptions(t *testing.T) {
	t.Run("config with options", func(t *testing.T) {
		server := grpcgateway.NewWithConfigOptions[transaction.Tx](
			func(c *grpcgateway.Config) {
				c.Address = "localhost:9090"
				c.Enable = true
			},
		)

		cfg := server.Config().(*grpcgateway.Config)
		require.Equal(t, "localhost:9090", cfg.Address)
		require.True(t, cfg.Enable)
	})
}
