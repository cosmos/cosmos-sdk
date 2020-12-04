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
	epochNumber := k.sk.GetEpochNumber(ctx)
	// execute all epoch actions
	iterator := k.GetEpochActionsIteratorByEpochNumber(ctx, epochNumber)

	for ; iterator.Valid(); iterator.Next() {
		msg := k.GetEpochActionByIterator(iterator)

		switch msg := msg.(type) {
		case *types.MsgUnjail:
			k.EpochUnjail(ctx, msg)
		default:
			panic(fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
		// dequeue processed item
		k.DeleteByKey(ctx, iterator.Key())
	}
}
