package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

// NewHandler defines the IBC handler
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case client.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, k.ClientKeeper, msg)

		case client.MsgUpdateClient:
			return client.HandleMsgUpdateClient(ctx, k.ClientKeeper, msg)

		case client.MsgSubmitMisbehaviour:
			return client.HandleMsgSubmitMisbehaviour(ctx, k.ClientKeeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized IBC Client message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}
