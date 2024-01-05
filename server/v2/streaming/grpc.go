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

// ListenEndBlock listens to end block request and responses.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
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

// ListenCommit listens to commit responses and state changes for the current block.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
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
