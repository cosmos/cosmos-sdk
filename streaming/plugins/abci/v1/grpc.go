package v1

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ baseapp.ABCIListener = (*GRPCClient)(nil)

// GRPCClient is an implementation of the ABCIListener interface that talks over RPC.
type GRPCClient struct {
	client ABCIListenerServiceClient
}

func (m GRPCClient) ListenBeginBlock(ctx context.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error {
	_, err := m.client.ListenBeginBlock(ctx, &ListenBeginBlockRequest{
		Req: &req,
		Res: &res,
	})
	return err
}

func (m GRPCClient) ListenEndBlock(ctx context.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error {
	_, err := m.client.ListenEndBlock(ctx, &ListenEndBlockRequest{
		Req: &req,
		Res: &res,
	})
	return err
}

func (m GRPCClient) ListenDeliverTx(goCtx context.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := m.client.ListenDeliverTx(ctx, &ListenDeliverTxRequest{
		BlockHeight: ctx.BlockHeight(),
		Req:         &req,
		Res:         &res,
	})
	return err
}

func (m GRPCClient) ListenCommit(goCtx context.Context, res abci.ResponseCommit, changeSet []*store.StoreKVPair) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := m.client.ListenCommit(ctx, &ListenCommitRequest{
		BlockHeight: ctx.BlockHeight(),
		Res:         &res,
		ChangeSet:   changeSet,
	})
	return err
}

// GRPCServer is the gRPC server that GRPCClient talks to.
type GRPCServer struct {
	// This is the real implementation
	Impl baseapp.ABCIListener
}

func (m GRPCServer) ListenBeginBlock(ctx context.Context, request *ListenBeginBlockRequest) (*Empty, error) {
	if err := m.Impl.ListenBeginBlock(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (m GRPCServer) ListenEndBlock(ctx context.Context, request *ListenEndBlockRequest) (*Empty, error) {
	if err := m.Impl.ListenEndBlock(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (m GRPCServer) ListenDeliverTx(ctx context.Context, request *ListenDeliverTxRequest) (*Empty, error) {
	if err := m.Impl.ListenDeliverTx(ctx, *request.Req, *request.Res); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}

func (m GRPCServer) ListenCommit(ctx context.Context, request *ListenCommitRequest) (*Empty, error) {
	if err := m.Impl.ListenCommit(ctx, *request.Res, request.ChangeSet); err != nil {
		return nil, err
	}
	return &Empty{}, nil
}
