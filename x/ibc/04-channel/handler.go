package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// HandleMsgChannelOpenInit defines the sdk.Handler for MsgChannelOpenInit
func HandleMsgChannelOpenInit(ctx sdk.Context, k keeper.Keeper, portCap *capability.Capability, msg types.MsgChannelOpenInit) (*sdk.Result, *capability.Capability, error) {
	capKey, err := k.ChanOpenInit(
		ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version,
	)
	if err != nil {
		return nil, nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]), // TODO: iterate
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, capKey, nil
}

// HandleMsgChannelOpenTry defines the sdk.Handler for MsgChannelOpenTry
func HandleMsgChannelOpenTry(ctx sdk.Context, k keeper.Keeper, portCap *capability.Capability, msg types.MsgChannelOpenTry) (*sdk.Result, *capability.Capability, error) {
	proofInit, err := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	capKey, err := k.ChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID,
		portCap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion, proofInit, msg.ProofHeight,
	)
	if err != nil {
		return nil, nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenTry,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
			sdk.NewAttribute(types.AttributeCounterpartyPortID, msg.Channel.Counterparty.PortID),
			sdk.NewAttribute(types.AttributeCounterpartyChannelID, msg.Channel.Counterparty.ChannelID),
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.Channel.ConnectionHops[0]), // TODO: iterate
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, capKey, nil
}

// HandleMsgChannelOpenAck defines the sdk.Handler for MsgChannelOpenAck
func HandleMsgChannelOpenAck(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg types.MsgChannelOpenAck) (*sdk.Result, error) {
	proofTry, err := commitmenttypes.UnpackAnyProof(&msg.ProofTry)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof try")
	}

	err = k.ChanOpenAck(
		ctx, msg.PortID, msg.ChannelID, channelCap, msg.CounterpartyVersion, proofTry, msg.ProofHeight,
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenAck,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelOpenConfirm defines the sdk.Handler for MsgChannelOpenConfirm
func HandleMsgChannelOpenConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg types.MsgChannelOpenConfirm) (*sdk.Result, error) {
	proofAck, err := commitmenttypes.UnpackAnyProof(&msg.ProofAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof ack")
	}

	err = k.ChanOpenConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, proofAck, msg.ProofHeight)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseInit defines the sdk.Handler for MsgChannelCloseInit
func HandleMsgChannelCloseInit(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg types.MsgChannelCloseInit) (*sdk.Result, error) {
	err := k.ChanCloseInit(ctx, msg.PortID, msg.ChannelID, channelCap)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseInit,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}

// HandleMsgChannelCloseConfirm defines the sdk.Handler for MsgChannelCloseConfirm
func HandleMsgChannelCloseConfirm(ctx sdk.Context, k keeper.Keeper, channelCap *capability.Capability, msg types.MsgChannelCloseConfirm) (*sdk.Result, error) {
	proofInit, err := commitmenttypes.UnpackAnyProof(&msg.ProofInit)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "invalid proof init")
	}

	err = k.ChanCloseConfirm(ctx, msg.PortID, msg.ChannelID, channelCap, proofInit, msg.ProofHeight)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeChannelCloseConfirm,
			sdk.NewAttribute(types.AttributeKeyPortID, msg.PortID),
			sdk.NewAttribute(types.AttributeKeyChannelID, msg.ChannelID),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	})

	return &sdk.Result{
		Events: ctx.EventManager().Events().ToABCIEvents(),
	}, nil
}
