package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var _ types.QueryServer = Keeper{}

// CurrentPlan implements the Query/CurrentPlan gRPC method
func (k Keeper) CurrentPlan(c context.Context, _ *types.QueryCurrentPlanRequest) (*types.QueryCurrentPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	plan, found := k.GetUpgradePlan(ctx)
	if !found {
		return &types.QueryCurrentPlanResponse{}, nil
	}

	return &types.QueryCurrentPlanResponse{Plan: &plan}, nil
}

// AppliedPlan implements the Query/AppliedPlan gRPC method
func (k Keeper) AppliedPlan(c context.Context, req *types.QueryAppliedPlanRequest) (*types.QueryAppliedPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	applied := k.GetDoneHeight(ctx, req.Name)
	if applied == 0 {
		return &types.QueryAppliedPlanResponse{}, nil
	}

	return &types.QueryAppliedPlanResponse{Height: applied}, nil
}

// UpgradedConsensusState implements the Query/UpgradedConsensusState gRPC method
func (k Keeper) UpgradedConsensusState(c context.Context, req *types.QueryUpgradedConsensusStateRequest) (*types.QueryUpgradedConsensusStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	consState, found := k.GetUpgradedConsensusState(ctx, req.LastHeight)
	if !found {
		return &types.QueryUpgradedConsensusStateResponse{}, nil
	}

	return &types.QueryUpgradedConsensusStateResponse{
		UpgradedConsensusState: consState,
	}, nil
}
