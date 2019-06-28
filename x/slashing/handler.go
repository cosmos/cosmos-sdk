package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgUnjail:
			return handleMsgUnjail(ctx, msg, k)

		default:
			errMsg := fmt.Sprintf("unrecognized slashing message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func handleMsgUnjail(ctx sdk.Context, msg MsgUnjail, k Keeper) sdk.Result {
	validator := k.sk.Validator(ctx, msg.ValidatorAddr)
	if validator == nil {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	// cannot be unjailed if no self-delegation exists
	selfDel := k.sk.Delegation(ctx, sdk.AccAddress(msg.ValidatorAddr), msg.ValidatorAddr)
	if selfDel == nil {
		return ErrMissingSelfDelegation(k.codespace).Result()
	}

	if validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
		return ErrSelfDelegationTooLowToUnjail(k.codespace).Result()
	}

	// cannot be unjailed if not jailed
	if !validator.IsJailed() {
		return ErrValidatorNotJailed(k.codespace).Result()
	}

	consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())

	info, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	// cannot be unjailed if tombstoned
	if info.Tombstoned {
		return ErrValidatorJailed(k.codespace).Result()
	}

	// cannot be unjailed until out of jail
	if ctx.BlockHeader().Time.Before(info.JailedUntil) {
		return ErrValidatorJailed(k.codespace).Result()
	}

	k.sk.Unjail(ctx, consAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}
