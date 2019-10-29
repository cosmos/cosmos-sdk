package mockbank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgRecvPacket:
			return handleMsgRecvPacket(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("failed to parse message").Result()
		}
	}
}

// handleMsgRecvPacket defines the sdk.Handler for MsgRecvPacket
func handleMsgRecvPacket(ctx sdk.Context, k Keeper, msg MsgRecvPacket) (res sdk.Result) {
	err := k.ReceivePacket(ctx, msg.Packet, msg.Proofs[0], msg.Height)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}
