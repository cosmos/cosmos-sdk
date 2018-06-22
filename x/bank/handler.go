package bank

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handler handles "bank" type messages
type Handler struct {
	k Keeper
}

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return Handler{k}
}

// Implements sdk.Handler
func (h Handler) Handle(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgSend:
		return handleMsgSend(ctx, h.k, msg)
	case MsgIssue:
		return handleMsgIssue(ctx, h.k, msg)
	default:
		errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

// Implements sdk.Handler
func (h Handler) Type() string {
	return MsgType
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
	panic("not implemented yet")
}
