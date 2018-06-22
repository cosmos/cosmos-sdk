package simplestake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handler handlers "simplestake" type messages
type Handler struct {
	k Keeper
}

// NewHandler returns a handler for "simplestake" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return Handler{k}
}

// Implements sdk.Handler
func (h Handler) Handle(ctx sdk.Context, msg sdk.Msg) sdk.Result {
	switch msg := msg.(type) {
	case MsgBond:
		return handleMsgBond()
	case MsgUnbond:
		return handleMsgUnbond()
	default:
		return sdk.ErrUnknownRequest("No match for message type.").Result()
	}
}

// Implements sdk.Handler
func (h Handler) Type() string {
	return moduleName
}

func handleMsgBond(ctx sdk.Context, k Keeper, msg MsgBond) sdk.Result {
	// Removed ValidatorSet from result because it does not get used.
	// TODO: Implement correct bond/unbond handling
	return sdk.Result{
		Code: sdk.ABCICodeOK,
	}
}

func handleMsgUnbond() sdk.Result {
	return sdk.Result{
		Code: sdk.ABCICodeOK,
	}
}
