package streaming

import (
	"context"
	"os"

	"github.com/hashicorp/go-plugin"
)

var _ Listener = (*GRPCClient)(nil)

// GRPCClient is an implementation of the ABCIListener interface that talks over RPC.
type GRPCClient struct {
	client ListenerServiceClient
}

// ListenDeliverBlock listens for block delivery requests and responses.
// It retrieves a types.Context from the provided context.Context.
// If the node is configured to stop on listening errors, it will terminate
// and exit with a non-zero code upon encountering an error.
//
// Panics if a types.Context is not properly attached to the provided context.Context.
func (m *GRPCClient) ListenDeliverBlock(goCtx context.Context, req ListenDeliverBlockRequest) error {
	ctx := goCtx.(Context)
	sm := ctx.StreamingManager()
	_, err := m.client.ListenDeliverBlock(goCtx, &req)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("DeliverBLock listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

// ListenStateChanges listens for state changes in the current block.
// It retrieves a types.Context from the provided context.Context.
// If the node is configured to stop on listening errors, it will terminate
// and exit with a non-zero code upon encountering an error.
//
// Panics if a types.Context is not properly attached to the provided context.Context.
func (m *GRPCClient) ListenStateChanges(goCtx context.Context, changeSet []*StoreKVPair) error {
	ctx := goCtx.(Context)
	sm := ctx.StreamingManager()
	request := &ListenStateChangesRequest{BlockHeight: ctx.BlockHeight(), ChangeSet: changeSet}
	_, err := m.client.ListenStateChanges(goCtx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("Commit listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

func cleanupAndExit() {
	plugin.CleanupClients()
	os.Exit(1)
}

var _ ListenerServiceServer = (*GRPCServer)(nil)

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl Listener
}

func (m GRPCServer) ListenDeliverBlock(ctx context.Context, request *ListenDeliverBlockRequest) (*ListenDeliverBlockResponse, error) {
	if err := m.Impl.ListenDeliverBlock(ctx, *request); err != nil {
		return nil, err
	}
	return &ListenDeliverBlockResponse{}, nil
}

func (m GRPCServer) ListenStateChanges(ctx context.Context, request *ListenStateChangesRequest) (*ListenStateChangesResponse, error) {
	if err := m.Impl.ListenStateChanges(ctx, request.ChangeSet); err != nil {
		return nil, err
	}
	return &ListenStateChangesResponse{}, nil
}
