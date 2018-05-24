package baseaccount

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// NewHandler returns a handler for "baseaccount" type messages.
func NewHandler(am auth.AccountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgChangeKey:
			return handleMsgChangeKey(ctx, am, msg)
		default:
			errMsg := "Unrecognized baseaccount Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgChangeKey
// Should be very expensive, because once this happens, an account is un-prunable
func handleMsgChangeKey(ctx sdk.Context, am auth.AccountMapper, msg MsgChangeKey) sdk.Result {

	err := am.SetPubKey(ctx, msg.Address, msg.NewPubKey)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{
		Tags: sdk.NewTags("action", []byte("changePubkey"), "address", msg.Address.Bytes(), "pubkey", msg.NewPubKey.Bytes()),
	}
}
