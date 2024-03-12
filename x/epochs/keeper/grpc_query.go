package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/epochs keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

// NewQuerier initializes new querier.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// EpochInfos provide running epochInfos.
func (q Querier) EpochInfos(c context.Context, _ *types.QueryEpochsInfoRequest) (*types.QueryEpochsInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	epochs, err := q.Keeper.AllEpochInfos(ctx)
	return &types.QueryEpochsInfoResponse{
		Epochs: epochs,
	}, err
}

// CurrentEpoch provides current epoch of specified identifier.
func (q Querier) CurrentEpoch(c context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Identifier == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier is empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	info := q.Keeper.GetEpochInfo(ctx, req.Identifier)
	if info.Identifier != req.Identifier {
		return nil, errors.New("not available identifier")
	}

	return &types.QueryCurrentEpochResponse{
		CurrentEpoch: info.CurrentEpoch,
	}, nil
}
