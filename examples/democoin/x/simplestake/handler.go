package simplestake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "simplestake" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgBond:
			return handleMsgBond(ctx, k, msg)
		case MsgUnbond:
			return handleMsgUnbond(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("No match for message type.").Result()
		}
	}
}

func handleMsgBond(ctx sdk.Context, k Keeper, msg MsgBond) sdk.Result {
	// Removed ValidatorSet from result because it does not get used.
	// TODO: Implement correct bond/unbond handling
	return sdk.Result{
		Code: sdk.ABCICodeOK,
	}
}

func handleMsgUnbond(ctx sdk.Context, k Keeper, msg MsgUnbond) sdk.Result {
	return sdk.Result{
		Code: sdk.ABCICodeOK,
	}
}
