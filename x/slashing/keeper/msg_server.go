package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

type msgServer struct {
	Keeper
	stakingKeeper types.StakingKeeper
}

// NewMsgServerImpl returns an implementation of the slashing MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper, stakingKeeper types.StakingKeeper) types.MsgServer {
	return &msgServer{
		Keeper:        keeper,
		stakingKeeper: stakingKeeper,
	}
}

var _ types.MsgServer = msgServer{}

// Unjail implements MsgServer.Unjail method.
// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func (k msgServer) Unjail(goCtx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Queue epoch action and move all the execution logic to Epoch execution

	epochNumber := k.stakingKeeper.GetEpochNumber(ctx)
	k.QueueMsgForEpoch(ctx, epochNumber, msg)

	cacheCtx, _ := ctx.CacheContext()
	cacheCtx = cacheCtx.WithBlockHeight(k.stakingKeeper.GetNextEpochHeight(ctx))
	cacheCtx = cacheCtx.WithBlockTime(k.stakingKeeper.GetNextEpochTime(ctx))
	err := k.ExecuteQueuedUnjail(cacheCtx, msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUnjailResponse{}, nil
}
