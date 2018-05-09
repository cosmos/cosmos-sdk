package bank

import (
	"reflect"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx bapp.Context, msg sdk.Msg) bapp.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, k, msg)
		case MsgIssue:
			return handleMsgIssue(ctx, k, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx bapp.Context, k Keeper, msg MsgSend) bapp.Result {
	// NOTE: totalIn == totalOut should already have been checked

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	// TODO: add some tags so we can search it!
	return bapp.Result{} // TODO
}

// Handle MsgIssue.
func handleMsgIssue(ctx bapp.Context, k Keeper, msg MsgIssue) bapp.Result {
	panic("not implemented yet")
}
