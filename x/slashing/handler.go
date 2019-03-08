package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/tags"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgUnjail:
			return handleMsgUnjail(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

// Validators must submit a transaction to unjail itself after
// having been jailed (and thus unbonded) for downtime
func handleMsgUnjail(ctx sdk.Context, msg MsgUnjail, k Keeper) sdk.Result {
	validator := k.validatorSet.Validator(ctx, msg.ValidatorAddr)
	if validator == nil {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	// cannot be unjailed if no self-delegation exists
	selfDel := k.validatorSet.Delegation(ctx, sdk.AccAddress(msg.ValidatorAddr), msg.ValidatorAddr)
	if selfDel == nil {
		return ErrMissingSelfDelegation(k.codespace).Result()
	}

	if validator.ShareTokens(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
		return ErrSelfDelegationTooLowToUnjail(k.codespace).Result()
	}

	// cannot be unjailed if not jailed
	if !validator.GetJailed() {
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

	// unjail the validator
	k.validatorSet.Unjail(ctx, consAddr)

	tags := sdk.NewTags(
		tags.Action, tags.ActionValidatorUnjailed,
		tags.Validator, msg.ValidatorAddr.String(),
	)

	return sdk.Result{
		Tags: tags,
	}
}
