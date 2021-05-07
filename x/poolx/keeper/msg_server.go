package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
)

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

type msgServer struct {
	Keeper
}

func (msgServer) CreatePool(c context.Context, _ *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_ = ctx

	// TODO(levi) delegate work to the keeper -- ie really create the pool

	// TODO(levi) emit events (& learn conventions)
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		sdk.EventTypeMessage,
	// 		// ???
	// 	),
	// )

	return &types.MsgCreatePoolResponse{}, nil
}

func (msgServer) FundPool(c context.Context, _ *types.MsgFundPool) (*types.MsgFundPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	_ = ctx

	// TODO(levi) delegate work to the keeper -- ie really fund the pool

	// TODO(levi) emit events (& learn conventions)
	// ctx.EventManager().EmitEvent(
	// 	sdk.NewEvent(
	// 		sdk.EventTypeMessage,
	// 		// ???
	// 	),
	// )

	return &types.MsgFundPoolResponse{}, nil
}
