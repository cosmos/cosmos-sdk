package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns sdk.Handler for IBC token transfer module messages
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer message type: %T", msg)
		}
	}
}

// See createOutgoingPacket in spec:https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleMsgTransfer(ctx sdk.Context, k Keeper, msg MsgTransfer) (*sdk.Result, error) {
	if err := k.SendTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.DestinationHeight, msg.Amount, msg.Sender, msg.Receiver,
	); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("IBC transfer: %s from %s to %s", msg.Amount, msg.Sender, msg.Receiver)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(AttributeKeyReceiver, msg.Receiver),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
