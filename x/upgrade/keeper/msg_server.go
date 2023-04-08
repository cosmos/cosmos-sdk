package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the upgrade MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

var _ types.MsgServer = msgServer{}

// SoftwareUpgrade implements the Msg/SoftwareUpgrade Msg service.
func (k msgServer) SoftwareUpgrade(goCtx context.Context, req *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, errors.Wrap(err, "authority")
	}

	if k.authority != req.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	if err := req.Plan.ValidateBasic(); err != nil {
		return nil, errors.Wrap(err, "plan")
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
	if _, err := sdk.AccAddressFromBech32(req.Authority); err != nil {
		return nil, errors.Wrap(err, "authority")
	}

	if k.authority != req.Authority {
		return nil, errors.Wrapf(gov.ErrInvalidSigner, "expected %s got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.ClearUpgradePlan(ctx)

	return &types.MsgCancelUpgradeResponse{}, nil
}
