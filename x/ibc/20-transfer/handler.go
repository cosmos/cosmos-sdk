package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case ibc.MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		case ibc.MsgPacket:
			switch packet := msg.Packet.Data.(type) {
			case PacketDataTransfer:
				return handlePacketTransfer(ctx, k, packet)
			default:
				// TODO: source chain sent wrong packet, shutdown channel
				return sdk.Result{}
			}
		default:
			// XXX: unknown
			return sdk.Result{}
		}
	}
}

// handleMsgTransfer defines the sdk.Handler for MsgTransfer
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
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}

// handlePacketTransfer defines the sdk.Handler for PacketTransfer
func handleMsgRecvPacket(ctx sdk.Context, k Keeper, packet PacketTransfer) (res sdk.Result) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Signer.String()),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}
}
