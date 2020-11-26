package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the slashing MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Unjail implements MsgServer.Unjail method.
// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func (k msgServer) Unjail(goCtx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Queue epoch action and move all the execution logic to Epoch execution
	k.SaveEpochAction(ctx, 0, msg)
	return &types.MsgUnjailResponse{}, nil
}
