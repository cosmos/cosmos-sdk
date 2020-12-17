package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// EpochUnjail logic is moved from msgServer.Unjail
func (k Keeper) EpochUnjail(ctx sdk.Context, msg *types.MsgUnjail) error {
	valAddr, valErr := sdk.ValAddressFromBech32(msg.ValidatorAddr)
	if valErr != nil {
		return valErr
	}
	err := k.Unjail(ctx, valAddr)
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr),
		),
	)

	return nil
}

// ExecuteEpoch execute epoch actions
func (k Keeper) ExecuteEpoch(ctx sdk.Context) {
	// execute all epoch actions
	for iterator := k.GetEpochActionsIterator(ctx); iterator.Valid(); iterator.Next() {
		msg := k.GetEpochActionByIterator(iterator)

		switch msg := msg.(type) {
		case *types.MsgUnjail:
			cacheCtx, writeCache := ctx.CacheContext()
			err := k.EpochUnjail(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else {
				// TODO: report somewhere for logging edit not success or panic
				// panic(fmt.Sprintf("not be able to execute, %T", msg))
			}
		default:
			panic(fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
		// dequeue processed item
		k.DeleteByKey(ctx, iterator.Key())
	}
}
