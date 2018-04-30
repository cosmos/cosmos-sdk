package auth

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "auth" type messages.
func NewHandler(am accountMapper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgClaimAccount:
			return handleMsgClaimAccount(ctx, am, msg)
		case MsgChangeKey:
			return handleMsgChangeKey(ctx, am, msg)
		default:
			errMsg := "Unrecognized auth Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgClaimAccount
func handleMsgClaimAccount(ctx sdk.Context, am sdk.AccountMapper, msg MsgClaimAccount) sdk.Result {
	// PubKey matches Address already checked in msg.ValidateBasic

	acc := am.GetAccount(ctx, msg.Address)
	if acc == nil {
		return sdk.ErrUnknownAddress(msg.Address.String()).Result()
	}
	if len(acc.GetPubKey().Bytes()) == 0 {
		return sdk.ErrInvalidPubKey("Account already claimed").Result()
	}
	acc.SetPubKey(msg.PubKey)
	am.SetAccount(ctx, acc)

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}

// Handle MsgChangeKey
// Should be very very expensive, because once this happens, an account is un-prunable
func handleMsgChangeKey(ctx sdk.Context, am sdk.AccountMapper, msg MsgChangeKey) sdk.Result {

	acc := am.GetAccount(ctx, msg.Address)
	if acc == nil {
		return sdk.ErrUnknownAddress(msg.Address.String()).Result()
	}
	acc.SetPubKey(msg.NewPubKey)
	am.SetAccount(ctx, acc)

	// TODO: add some tags so we can search it!
	return sdk.Result{} // TODO
}
