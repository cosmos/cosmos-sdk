package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, k, msg)
		case MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)
		default:
			errMsg := "Unrecognized bank Msg type: %s" + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k Keeper, msg MsgSend) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return ErrSendDisabled(k.Codespace()).Result()
	}
	tags, err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k Keeper, msg MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		return ErrSendDisabled(k.Codespace()).Result()
	}
	tags, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}
