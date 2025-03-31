package keeper

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

// CommunityPool queries the community pool coins
func (k Querier) CommunityPool(ctx context.Context, req *types.QueryCommunityPoolRequest) (*types.QueryCommunityPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	amount, err := k.GetCommunityPool(sdkCtx)
	if err != nil {
		return nil, err
	}
	return &types.QueryCommunityPoolResponse{Pool: amount}, nil
}

// ContinuousFund queries a continuous fund by its recipient address.
func (k Querier) ContinuousFund(ctx context.Context, req *types.QueryContinuousFundRequest) (*types.QueryContinuousFundResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	acc, err := k.authKeeper.AddressCodec().StringToBytes(req.Recipient)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("invalid address: %w", err).Error())
	}

	fund, err := k.Keeper.ContinuousFunds.Get(sdkCtx, acc)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("not found %s", req.Recipient))
	}

	return &types.QueryContinuousFundResponse{ContinuousFund: fund}, nil
}

// ContinuousFunds queries all continuous funds in the store.
func (k Querier) ContinuousFunds(ctx context.Context, req *types.QueryContinuousFundsRequest) (*types.QueryContinuousFundsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	funds, err := k.GetAllContinuousFunds(sdkCtx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("failed to fetch continuous funds: %w", err).Error())
	}

	return &types.QueryContinuousFundsResponse{ContinuousFunds: funds}, nil
}

// Params queries params of x/protocolpool module.
func (k Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	params, err := k.Keeper.Params.Get(sdkCtx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "params not found")
		}
		return nil, err
	}

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}
