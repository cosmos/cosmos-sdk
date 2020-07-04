package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// Params queries params of distribution module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ValidatorOutstandingRewards queries rewards of a validator address
func (k Keeper) ValidatorOutstandingRewards(c context.Context, req *types.QueryValidatorOutstandingRewardsRequest) (*types.QueryValidatorOutstandingRewardsResponse, error) {
	if req.String() == "" || req.ValidatorAddress.Empty() {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var rewards types.ValidatorOutstandingRewards

	store := ctx.KVStore(k.storeKey)
	rewardsStore := prefix.NewStore(store, types.GetValidatorOutstandingRewardsKey(req.ValidatorAddress))

	res, err := query.Paginate(rewardsStore, req.Req, func(key []byte, value []byte) error {
		err := k.cdc.UnmarshalBinaryBare(value, &rewards)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return &types.QueryValidatorOutstandingRewardsResponse{}, nil
	}

	return &types.QueryValidatorOutstandingRewardsResponse{Rewards: rewards, Res: res}, nil
}
