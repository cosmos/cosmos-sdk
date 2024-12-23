package store

import (
	"testing"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		config    server.ConfigMap
		expectErr bool
	}{
		{
			name: "valid config",
			config: server.ConfigMap{
				"store": map[string]any{
					"app-db-backend": "goleveldb",
				},
				"home": "/tmp",
			},
			expectErr: false,
		},
		{
			name: "invalid config type",
			config: server.ConfigMap{
				"store": "invalid",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := New[transaction.Tx](nil, tt.config)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, srv)
			require.Equal(t, ServerName, srv.Name())
		})
	}
}
