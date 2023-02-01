package v1

import (
	"context"
	"os"

	"github.com/hashicorp/go-plugin"
	abci "github.com/tendermint/tendermint/abci/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ storetypes.ABCIListener = (*GRPCClient)(nil)

// GRPCClient is an implementation of the ABCIListener interface that talks over RPC.
type GRPCClient struct {
	client ABCIListenerServiceClient
}

func (m *GRPCClient) ListenBeginBlock(goCtx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sm := ctx.StreamingManager()
	request := &ListenBeginBlockRequest{Req: &req, Res: &res}
	_, err := m.client.ListenBeginBlock(ctx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("BeginBlock listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

func (m *GRPCClient) ListenEndBlock(goCtx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sm := ctx.StreamingManager()
	request := &ListenEndBlockRequest{Req: &req, Res: &res}
	_, err := m.client.ListenEndBlock(ctx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("EndBlock listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

func (m *GRPCClient) ListenDeliverTx(goCtx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sm := ctx.StreamingManager()
	request := &ListenDeliverTxRequest{BlockHeight: ctx.BlockHeight(), Req: &req, Res: &res}
	_, err := m.client.ListenDeliverTx(ctx, request)
	if err != nil && sm.StopNodeOnErr {
		ctx.Logger().Error("DeliverTx listening hook failed", "height", ctx.BlockHeight(), "err", err)
		cleanupAndExit()
	}
	return err
}

func (m *GRPCClient) ListenCommit(goCtx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	sm := ctx.StreamingManager()
	request := &ListenCommitRequest{BlockHeight: ctx.BlockHeight(), Res: &res, ChangeSet: changeSet}
	_, err := m.client.ListenCommit(ctx, request)
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
