package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/epochs/types"
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
func (q Querier) EpochInfos(ctx context.Context, _ *types.QueryEpochInfosRequest) (*types.QueryEpochInfosResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	epochs, err := q.AllEpochInfos(sdkCtx)
	return &types.QueryEpochInfosResponse{
		Epochs: epochs,
	}, err
}

// CurrentEpoch provides current epoch of specified identifier.
func (q Querier) CurrentEpoch(ctx context.Context, req *types.QueryCurrentEpochRequest) (*types.QueryCurrentEpochResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Identifier == "" {
		return nil, status.Error(codes.InvalidArgument, "identifier is empty")
	}

	info, err := q.EpochInfo.Get(ctx, req.Identifier)
	if err != nil {
		return nil, errors.New("not available identifier")
	}

	return &types.QueryCurrentEpochResponse{
		CurrentEpoch: info.CurrentEpoch,
	}, nil
}
