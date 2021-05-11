package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	xp "github.com/cosmos/cosmos-sdk/x/upgrade/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var _ types.QueryServer = Keeper{}

// CurrentPlan implements the Query/CurrentPlan gRPC method
func (k Keeper) CurrentPlan(c context.Context, req *types.QueryCurrentPlanRequest) (*types.QueryCurrentPlanResponse, error) {
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

// ModuleVersions implements the Query/QueryModuleVersions gRPC method
func (k Keeper) ModuleVersions(c context.Context, req *types.QueryModuleVersionsRequest) (*types.QueryModuleVersionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	// get version map from x/upgrade store
	vm := k.GetModuleVersionMap(ctx)

	// make response slice
	res := make([]*types.ModuleVersion, 0)

	// check if a specific module was requested
	if len(req.ModuleName) > 0 {
		// check if the requested module exists
		if version, found := vm[req.ModuleName]; found {
			// add the requested module
			res = append(res, &types.ModuleVersion{Name: req.ModuleName, Version: version})
		} else { // module was requested, but not found
			return &types.QueryModuleVersionsResponse{}, errors.Wrapf(errors.ErrNotFound, "x/upgrade: QueryModuleVersions")
		}
	} else {
		// if no module requested, add entire vm to slice
		for m, v := range vm {
			res = append(res, &types.ModuleVersion{Name: m, Version: v})
		}
	}

	res = xp.Sort(res)

	return &types.QueryModuleVersionsResponse{
		ModuleVersions: res,
	}, nil
}
