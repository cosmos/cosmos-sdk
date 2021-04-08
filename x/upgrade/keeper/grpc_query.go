package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
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

// VersionMap implements the Query/VersionMap gRPC method
func (k Keeper) VersionMap(c context.Context, req *types.QueryVersionMapRequest) (*types.QueryVersionMapResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	// get the version map from the upgrade store
	vm := k.GetModuleVersionMap(ctx)

	// check if a specific module was requested and if
	if len(req.ModuleName) > 0 {
		// that module exists in the version map
		if vm[req.GetModuleName()] > 0 {
			// make a versionmap containing only the requested module
			singleVM := make(module.VersionMap)
			singleVM[req.GetModuleName()] = vm[req.GetModuleName()]
			vm = singleVM
		} else { // the requested module was not found

			return &types.QueryVersionMapResponse{}, errors.Wrapf(errors.ErrNotFound, "x/upgrade QueryVersionMap")
		}
	}

	vma := make([]*types.ModuleConsensusVersion, len(vm))

	for m, v := range vm {
		vma = append(vma, &types.ModuleConsensusVersion{Module: m, Version: v})
	}

	return &types.QueryVersionMapResponse{
		VersionMap: vma,
	}, nil
}
