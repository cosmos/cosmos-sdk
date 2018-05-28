package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		// NOTE msg already has validate basic run
		switch msg := msg.(type) {
		case MsgUnrevoke:
			return handleMsgUnrevoke(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("invalid message parse in staking module").Result()
		}
	}
}

func handleMsgUnrevoke(ctx sdk.Context, msg MsgUnrevoke, k Keeper) sdk.Result {
	validator := k.stakeKeeper.Validator(ctx, msg.ValidatorAddr)
	if validator == nil {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	info, found := k.getValidatorSigningInfo(ctx, validator.GetPubKey().Address())
	if !found {
		return ErrNoValidatorForAddress(k.codespace).Result()
	}

	if ctx.BlockHeader().Time < info.JailedUntil {
		return ErrValidatorJailed(k.codespace).Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{}
	}

	k.stakeKeeper.Unrevoke(ctx, validator.GetPubKey())

	tags := sdk.NewTags("action", []byte("unrevoke"), "validator", msg.ValidatorAddr.Bytes())
	return sdk.Result{
		Tags: tags,
	}
}
