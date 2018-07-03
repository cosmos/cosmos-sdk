package simplestake

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "simplestake" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg.(type) {
		case MsgBond:
			return handleMsgBond()
		case MsgUnbond:
			return handleMsgUnbond()
		default:
			return sdk.ErrUnknownRequest("No match for message type.").Result()
		}
	}
}

func handleMsgBond() sdk.Result {
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
