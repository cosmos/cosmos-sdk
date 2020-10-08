package transfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

// NewHandler returns sdk.Handler for IBC token transfer module messages
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgTransfer:
			return handleMsgTransfer(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized ICS-20 transfer message type: %T", msg)
		}
	}
}

// See createOutgoingPacket in spec:https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func handleMsgTransfer(ctx sdk.Context, k keeper.Keeper, msg *types.MsgTransfer) (*sdk.Result, error) {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}
	if err := k.SendTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.Token, sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp,
	); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("IBC fungible token transfer", "token", msg.Token.Denom, "amount", msg.Token.Amount.String(), "sender", msg.Sender, "receiver", msg.Receiver)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
