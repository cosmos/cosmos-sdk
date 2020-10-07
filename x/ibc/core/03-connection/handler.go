package connection

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

// HandleMsgConnectionOpenInit defines the sdk.Handler for MsgConnectionOpenInit
func HandleMsgConnectionOpenInit(ctx sdk.Context, k keeper.Keeper, msg *types.MsgConnectionOpenInit) (*sdk.Result, error) {
	if err := k.ConnOpenInit(
		ctx, msg.ConnectionId, msg.ClientId, msg.Counterparty, msg.Version,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open init failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenInit,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyConnectionID, msg.Counterparty.ConnectionId),
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

// HandleMsgConnectionOpenTry defines the sdk.Handler for MsgConnectionOpenTry
func HandleMsgConnectionOpenTry(ctx sdk.Context, k keeper.Keeper, msg *types.MsgConnectionOpenTry) (*sdk.Result, error) {
	targetClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "client in msg is not exported.ClientState. invalid client: %v.", targetClient)
	}

	if err := k.ConnOpenTry(
		ctx, msg.DesiredConnectionId, msg.CounterpartyChosenConnectionId, msg.Counterparty, msg.ClientId, targetClient,
		msg.CounterpartyVersions, msg.ProofInit, msg.ProofClient, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open try failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenTry,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.DesiredConnectionId),
			sdk.NewAttribute(types.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyConnectionID, msg.Counterparty.ConnectionId),
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

// HandleMsgConnectionOpenAck defines the sdk.Handler for MsgConnectionOpenAck
func HandleMsgConnectionOpenAck(ctx sdk.Context, k keeper.Keeper, msg *types.MsgConnectionOpenAck) (*sdk.Result, error) {
	targetClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "client in msg is not exported.ClientState. invalid client: %v", targetClient)
	}

	if err := k.ConnOpenAck(
		ctx, msg.ConnectionId, targetClient, msg.Version, msg.CounterpartyConnectionId,
		msg.ProofTry, msg.ProofClient, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open ack failed")
	}

	connectionEnd, _ := k.GetConnection(ctx, msg.ConnectionId)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenAck,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyClientID, connectionEnd.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, connectionEnd.Counterparty.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyConnectionID, connectionEnd.Counterparty.ConnectionId),
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

// HandleMsgConnectionOpenConfirm defines the sdk.Handler for MsgConnectionOpenConfirm
func HandleMsgConnectionOpenConfirm(ctx sdk.Context, k keeper.Keeper, msg *types.MsgConnectionOpenConfirm) (*sdk.Result, error) {
	if err := k.ConnOpenConfirm(
		ctx, msg.ConnectionId, msg.ProofAck, msg.ProofHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open confirm failed")
	}

	connectionEnd, _ := k.GetConnection(ctx, msg.ConnectionId)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeConnectionOpenConfirm,
			sdk.NewAttribute(types.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyClientID, connectionEnd.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyClientID, connectionEnd.Counterparty.ClientId),
			sdk.NewAttribute(types.AttributeKeyCounterpartyConnectionID, connectionEnd.Counterparty.ConnectionId),
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
