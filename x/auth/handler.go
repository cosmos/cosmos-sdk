package auth

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "auth" type messages.
func NewHandler(am AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgChangeKey:
			return handleMsgChangeKey(ctx, am, msg)
		default:
			errMsg := "Unrecognized auth Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgChangeKey
// Should be very expensive, because once this happens, an account is un-prunable
func handleMsgChangeKey(ctx sdk.Context, am AccountMapper, msg MsgChangeKey) sdk.Result {

	err := am.setPubKey(ctx, msg.Address, msg.NewPubKey)
	if err != nil {
		return err.Result()
	}

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}
