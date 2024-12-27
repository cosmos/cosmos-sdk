package grpc

import (
	"context"
	"testing"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestGetHeightFromCtx(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		wantHeight  uint64
		wantErr     bool
		errContains string
	}{
		{
			name:       "no metadata returns zero height",
			ctx:        context.Background(),
			wantHeight: 0,
			wantErr:    false,
		},
		{
			name: "valid height",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(BlockHeightHeader, "100"),
			),
			wantHeight: 100,
			wantErr:    false,
		},
		{
			name: "invalid height format",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(BlockHeightHeader, "invalid"),
			),
			wantHeight:  0,
			wantErr:     true,
			errContains: "unable to parse height",
		},
		{
			name: "multiple height values",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(BlockHeightHeader, "100", BlockHeightHeader, "200"),
			),
			wantHeight:  0,
			wantErr:     true,
			errContains: "must be of length 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			height, err := getHeightFromCtx(tt.ctx)
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.errContains)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if height != tt.wantHeight {
				t.Errorf("height = %v, want %v", height, tt.wantHeight)
			}
		})
	}
}

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfgOptions  []CfgOption
		wantAddress string
	}{
		{
			name:        "default config",
			cfgOptions:  nil,
			wantAddress: "localhost:9090",
		},
		{
			name: "custom address",
			cfgOptions: []CfgOption{
				func(cfg *Config) {
					cfg.Address = "localhost:8080"
				},
			},
			wantAddress: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewWithConfigOptions[transaction.Tx](tt.cfgOptions...)
			cfg := srv.Config().(*Config)
			if cfg.Address != tt.wantAddress {
				t.Errorf("address = %v, want %v", cfg.Address, tt.wantAddress)
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	logger := log.NewNopLogger()
	handlers := make(map[string]appmodulev2.Handler)
	queryable := func(ctx context.Context, version uint64, msg transaction.Msg) (transaction.Msg, error) {
		return msg, nil
	}

	srv, err := New[transaction.Tx](
		logger,
		nil,
		handlers,
		queryable,
		nil,
	)

	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if srv.Name() != ServerName {
		t.Errorf("server name = %v, want %v", srv.Name(), ServerName)
	}
}
