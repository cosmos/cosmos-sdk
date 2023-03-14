package abci

import (
	"context"
	"os"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/hashicorp/go-plugin"

	storetypes "cosmossdk.io/store/types"
)

var _ storetypes.ABCIListener = (*GRPCClient)(nil)

// GRPCClient is an implementation of the ABCIListener interface that talks over RPC.
type GRPCClient struct {
	client ABCIListenerServiceClient
}

// ListenBeginBlock listens to begin block request and responses.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
func (m *GRPCClient) ListenBeginBlock(goCtx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	ctx := goCtx.(storetypes.Context)
	sm := ctx.StreamingManager()
	request := &ListenBeginBlockRequest{Req: &req, Res: &res}
	_, err := m.client.ListenBeginBlock(goCtx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("BeginBlock listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

// ListenEndBlock listens to end block request and responses.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
func (m *GRPCClient) ListenEndBlock(goCtx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	ctx := goCtx.(storetypes.Context)
	sm := ctx.StreamingManager()
	request := &ListenEndBlockRequest{Req: &req, Res: &res}
	_, err := m.client.ListenEndBlock(goCtx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("EndBlock listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

// ListenDeliverTx listens to deliver tx request and responses.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
func (m *GRPCClient) ListenDeliverTx(goCtx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	ctx := goCtx.(storetypes.Context)
	sm := ctx.StreamingManager()
	request := &ListenDeliverTxRequest{BlockHeight: ctx.BlockHeight(), Req: &req, Res: &res}
	_, err := m.client.ListenDeliverTx(goCtx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("DeliverTx listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

// ListenCommit listens to commit responses and state changes for the current block.
// In addition, it retrieves a types.Context from a context.Context instance.
// It panics if a types.Context was not properly attached.
// When the node is configured to stop on listening errors,
// it will terminate immediately and exit with a non-zero code.
func (m *GRPCClient) ListenCommit(goCtx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	ctx := goCtx.(storetypes.Context)
	sm := ctx.StreamingManager()
	request := &ListenCommitRequest{BlockHeight: ctx.BlockHeight(), Res: &res, ChangeSet: changeSet}
	_, err := m.client.ListenCommit(goCtx, request)
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

var _ ABCIListenerServiceServer = (*GRPCServer)(nil)

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl storetypes.ABCIListener
}

func (m GRPCServer) ListenBeginBlock(ctx context.Context, request *ListenBeginBlockRequest) (*ListenBeginBlockResponse, error) {
	if err := m.Impl.ListenBeginBlock(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &ListenBeginBlockResponse{}, nil
}

func (m GRPCServer) ListenEndBlock(ctx context.Context, request *ListenEndBlockRequest) (*ListenEndBlockResponse, error) {
	if err := m.Impl.ListenEndBlock(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &ListenEndBlockResponse{}, nil
}

func (m GRPCServer) ListenDeliverTx(ctx context.Context, request *ListenDeliverTxRequest) (*ListenDeliverTxResponse, error) {
	if err := m.Impl.ListenDeliverTx(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &ListenDeliverTxResponse{}, nil
}

func (m GRPCServer) ListenCommit(ctx context.Context, request *ListenCommitRequest) (*ListenCommitResponse, error) {
	if err := m.Impl.ListenCommit(ctx, *request.Res, request.ChangeSet); err != nil {
		return nil, err
	}
	return &ListenCommitResponse{}, nil
}
