package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgRecvPacket:
			return handleMsgRecvPacket(ctx, k, msg)
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
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

// HandleMsgTransfer defines the sdk.Handler for MsgTransfer
func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (res sdk.Result) {
	err := k.SendTransfer(ctx, msg.SourcePort, msg.SourceChannel, msg.Amount, msg.Sender, msg.Receiver, msg.Source)
	if err != nil {
		return sdk.ResultFromError(err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver.String()),
		))

	return sdk.Result{Events: ctx.EventManager().Events()}
}
