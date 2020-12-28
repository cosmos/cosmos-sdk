package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// executeQueuedUnjailMsg logic is moved from msgServer.Unjail
func (k Keeper) executeQueuedUnjailMsg(ctx sdk.Context, msg *types.MsgUnjail) error {
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

func (k Keeper) executeQueuedSlashEvent(ctx sdk.Context, msg *types.SlashEvent) error {
	validator := k.sk.Validator(ctx, msg.Address)
	if validator == nil {
		return sdkerrors.Wrap(types.ErrBadValidatorAddr, msg.Address.String())
	}
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return err
	}
	k.sk.Slash(ctx, consAddr, msg.DistributionHeight, msg.ValidatorPower, msg.SlashPercent)
	return nil
}

// ExecuteEpoch execute epoch actions
func (k Keeper) ExecuteEpoch(ctx sdk.Context) {
	// execute all epoch actions
	for iterator := k.ek.GetEpochActionsIterator(ctx); iterator.Valid(); iterator.Next() {
		msg := k.ek.GetEpochActionByIterator(iterator)
		cacheCtx, writeCache := ctx.CacheContext()

		switch msg := msg.(type) {
		case *types.MsgUnjail:
			err := k.executeQueuedUnjailMsg(cacheCtx, msg)
			if err == nil {
				writeCache()
			} else {
				// TODO: report somewhere for logging edit not success or panic
				// panic(fmt.Sprintf("not be able to execute, %T", msg))
			}
		case *types.SlashEvent:
			err := k.executeQueuedSlashEvent(ctx, msg)
			if err == nil {
				writeCache()
			} else {
				// TODO: report somewhere for logging edit not success or panic
				panic(fmt.Sprintf("not be able to execute, %T, %s", msg, err.Error()))
			}
		default:
			panic(fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg))
		}
		// dequeue processed item
		k.ek.DeleteByKey(ctx, iterator.Key())
	}
}
