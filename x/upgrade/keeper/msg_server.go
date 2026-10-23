package keeper

import (
	"context"

	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type msgServer struct {
	*Keeper
}

// newMsgServerImpl returns an implementation of the upgrade MsgServer interface
// for the provided Keeper.
func newMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// NewMsgServerImpl returns an implementation of the upgrade MsgServer interface
// for the provided Keeper.
//
// Deprecated: register the service with RegisterMsgServer and run msgs from
// outside the module with msgservice.Dispatch. NewMsgServerImpl will be
// removed in the next release.
func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return newMsgServerImpl(k)
}

// RegisterMsgServer registers the x/upgrade Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, k *Keeper) {
	types.RegisterMsgServer(r, newMsgServerImpl(k))
}

var (
	_    types.MsgServer = msgServer{}
	_, _ sdk.Msg         = &types.MsgSoftwareUpgrade{}, &types.MsgCancelUpgrade{}
)

// SoftwareUpgrade implements the Msg/SoftwareUpgrade Msg service.
func (k msgServer) SoftwareUpgrade(goCtx context.Context, msg *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := sdk.ValidateAuthority(ctx, k.authority, msg.Authority); err != nil {
		return nil, err
	}

	err := k.ScheduleUpgrade(ctx, msg.Plan)
	if err != nil {
		return nil, err
	}

	return &types.MsgSoftwareUpgradeResponse{}, nil
}

// CancelUpgrade implements the Msg/CancelUpgrade Msg service.
func (k msgServer) CancelUpgrade(goCtx context.Context, msg *types.MsgCancelUpgrade) (*types.MsgCancelUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := sdk.ValidateAuthority(ctx, k.authority, msg.Authority); err != nil {
		return nil, err
	}

	err := k.ClearUpgradePlan(ctx)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelUpgradeResponse{}, nil
}
