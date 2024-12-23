package streaming

import (
	"context"
	"errors"
	"testing"

	coretesting "cosmossdk.io/core/testing"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"
)

func TestGRPCClient_ListenDeliverBlock(t *testing.T) {
	logger := coretesting.NewNopLogger()
	tests := []struct {
		name      string
		stopOnErr bool
		wantErr   bool
	}{
		{
			name:      "success - normal operation",
			stopOnErr: false,
			wantErr:   false,
		},
		{
			name:      "error - with StopNodeOnErr false",
			stopOnErr: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamingService := Manager{
				StopNodeOnErr: tt.stopOnErr,
			}
			mockCtx := NewMockContext(100, logger, streamingService)

			client := &GRPCClient{
				client: &MockListenerServiceClient{
					shouldError: tt.wantErr,
				},
			}
			err := client.ListenDeliverBlock(mockCtx, ListenDeliverBlockRequest{})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

var _ ListenerServiceClient = (*MockListenerServiceClient)(nil)

type MockListenerServiceClient struct {
	shouldError bool
}

func (m *MockListenerServiceClient) ListenDeliverBlock(ctx context.Context, req *ListenDeliverBlockRequest, opts ...grpc.CallOption) (*ListenDeliverBlockResponse, error) {
	if m.shouldError {
		return nil, errors.New("test error")
	}
	return &ListenDeliverBlockResponse{}, nil
}

func (m *MockListenerServiceClient) ListenStateChanges(ctx context.Context, req *ListenStateChangesRequest, opts ...grpc.CallOption) (*ListenStateChangesResponse, error) {
	if m.shouldError {
		return nil, errors.New("test error")
	}
	return &ListenStateChangesResponse{}, nil
}

func TestGRPCClient_ListenStateChanges(t *testing.T) {
	logger := coretesting.NewNopLogger()
	tests := []struct {
		name      string
		stopOnErr bool
		wantErr   bool
	}{
		{
			name:      "success - normal operation",
			stopOnErr: false,
			wantErr:   false,
		},
		{
			name:      "error - with StopNodeOnErr false",
			stopOnErr: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streamingService := Manager{
				StopNodeOnErr: tt.stopOnErr,
			}
			mockCtx := NewMockContext(100, logger, streamingService)

			client := &GRPCClient{
				client: &MockListenerServiceClient{
					shouldError: tt.wantErr,
				},
			}

			changeSet := []*StoreKVPair{
				{
					Key:   []byte("test-key"),
					Value: []byte("test-value"),
				},
			}
			err := client.ListenStateChanges(mockCtx, changeSet)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
