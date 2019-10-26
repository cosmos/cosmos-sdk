package mockbank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgRecvTransferPacket:
			return handleMsgRecvTransferPacket(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("failed to parse message").Result()
		}
	}
}

func handleMsgRecvTransferPacket(ctx sdk.Context, k Keeper, msg MsgRecvTransferPacket) (res sdk.Result) {
	err := k.ReceiveTransferPacket(ctx, msg.Packet, msg.Proofs[0], msg.Height)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}
