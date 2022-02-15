package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the upgrade MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

var _ types.MsgServer = msgServer{}

// SoftwareUpgrade implements the Msg/SoftwareUpgrade Msg service.
func (k msgServer) SoftwareUpgrade(goCtx context.Context, req *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	govAcct := k.authKeeper.GetModuleAddress(gov.ModuleName).String()
	if govAcct != req.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", govAcct, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ScheduleUpgrade(ctx, req.Plan)
	if err != nil {
		return nil, err
	}

	return &types.MsgSoftwareUpgradeResponse{}, nil
}

// CancelUpgrade implements the Msg/CancelUpgrade Msg service.
func (k msgServer) CancelUpgrade(goCtx context.Context, req *types.MsgCancelUpgrade) (*types.MsgCancelUpgradeResponse, error) {
	govAcct := k.authKeeper.GetModuleAddress(gov.ModuleName).String()
	if govAcct != req.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", govAcct, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)

	// clear any old IBC state stored by previous plan
	_, found := k.GetUpgradePlan(ctx)
	if !found {
		return nil, errors.Wrapf(errors.ErrNotFound, "no plan is currently scheduled")
	}

	store.Delete(types.PlanKey())

	return &types.MsgCancelUpgradeResponse{}, nil
}
