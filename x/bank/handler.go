package bank

import (
	"bytes"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
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
func handleMsgSend(ctx sdk.Context, k Keeper, msg MsgSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked

	tags, err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: tags,
	}
}

// Handle MsgIssue.
func handleMsgIssue(ctx sdk.Context, k Keeper, msg MsgIssue) sdk.Result {
	banker := k.GetBanker()
	tags := sdk.EmptyTags()

	// No banker was set and hence it's not possible to issue coins.
	if banker == nil {
		return ErrNoBanker(DefaultCodespace).Result()
	}

	if !bytes.Equal(banker, msg.Banker) {
		return ErrInvalidBanker(DefaultCodespace).Result()
	}

	for _, out := range msg.Outputs {
		_, subtags, err := k.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return err.Result()
		}
		tags = tags.AppendTags(subtags)
	}

	return sdk.Result{
		Tags: tags,
	}
}
