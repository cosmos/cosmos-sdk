package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ValidatorGet(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	validator, found := k.GetValidator(ctx, req.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound
	}

	return &types.QueryValidatorResponse{Validator: &validator}, nil
}
