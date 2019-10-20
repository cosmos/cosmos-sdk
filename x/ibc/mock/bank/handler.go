package mockbank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics04 "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		case MsgRecvTransferPacket:
			return handleMsgRecvTransferPacket(ctx, k, msg)
		default:
			return sdk.ErrUnknownRequest("failed to parse message").Result()
		}
	}
}

func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (res sdk.Result) {
	err := k.SendTransfer(ctx, msg.SrcPort, msg.SrcChannel, msg.Denomination, msg.Amount, msg.Sender, msg.Receiver, msg.Source)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, ics04.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
		))

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgRecvTransferPacket(ctx sdk.Context, k Keeper, msg MsgRecvTransferPacket) (res sdk.Result) {
	err := k.ReceiveTransfer(ctx, msg.Packet, msg.Proofs[0], msg.Height)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}
