package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

type MsgServer struct {
	Keeper
}

var _ clienttypes.MsgServer = MsgServer{}
var _ connectiontypes.MsgServer = MsgServer{}

func (k MsgServer) CreateClient(goCtx context.Context, msg *clienttypes.MsgCreateClient) (*clienttypes.MsgCreateClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}

	consensusState, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return nil, err
	}

	if err = k.ClientKeeper.CreateClient(ctx, msg.ClientId, clientState, consensusState); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			clienttypes.EventTypeCreateClient,
			sdk.NewAttribute(clienttypes.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(clienttypes.AttributeKeyClientType, clientState.ClientType()),
			sdk.NewAttribute(clienttypes.AttributeKeyConsensusHeight, clientState.GetLatestHeight().String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, clienttypes.AttributeValueCategory),
		),
	})

	return &clienttypes.MsgCreateClientResponse{}, nil
}

func (k MsgServer) UpdateClient(goCtx context.Context, msg *clienttypes.MsgUpdateClient) (*clienttypes.MsgUpdateClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	header, err := clienttypes.UnpackHeader(msg.Header)
	if err != nil {
		return nil, err
	}

	if err = k.ClientKeeper.UpdateClient(ctx, msg.ClientId, header); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, clienttypes.AttributeValueCategory),
		),
	)

	return &clienttypes.MsgUpdateClientResponse{}, nil
}

func (k MsgServer) UpgradeClient(goCtx context.Context, msg *clienttypes.MsgUpgradeClient) (*clienttypes.MsgUpgradeClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	upgradedClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}

	if err := upgradedClient.Validate(); err != nil {
		return nil, err
	}

	if err = k.ClientKeeper.UpgradeClient(ctx, msg.ClientId, upgradedClient, msg.UpgradeHeight, msg.ProofUpgrade); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, clienttypes.AttributeValueCategory),
		),
	)

	return &clienttypes.MsgUpgradeClientResponse{}, nil
}

func (k MsgServer) SubmitMisbehaviour(goCtx context.Context, msg *clienttypes.MsgSubmitMisbehaviour) (*clienttypes.MsgSubmitMisbehaviourResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	misbehaviour, err := clienttypes.UnpackMisbehaviour(msg.Misbehaviour)
	if err != nil {
		return nil, err
	}

	if err := k.ClientKeeper.CheckMisbehaviourAndUpdateState(ctx, misbehaviour); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to process misbehaviour for IBC client")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			clienttypes.EventTypeSubmitMisbehaviour,
			sdk.NewAttribute(clienttypes.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(clienttypes.AttributeKeyClientType, misbehaviour.ClientType()),
			sdk.NewAttribute(clienttypes.AttributeKeyConsensusHeight, misbehaviour.GetHeight().String()),
		),
	)

	return &clienttypes.MsgSubmitMisbehaviourResponse{}, nil
}

func (k MsgServer) ConnectionOpenInit(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenInit) (*connectiontypes.MsgConnectionOpenInitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.ConnectionKeeper.ConnOpenInit(
		ctx, msg.ConnectionId, msg.ClientId, msg.Counterparty, msg.Version,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open init failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenInit,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(connectiontypes.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyConnectionID, msg.Counterparty.ConnectionId),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, connectiontypes.AttributeValueCategory),
		),
	})

	return &connectiontypes.MsgConnectionOpenInitResponse{}, nil
}

func (k MsgServer) ConnectionOpenTry(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenTry) (*connectiontypes.MsgConnectionOpenTryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	targetClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "client in msg is not exported.ClientState. invalid client: %v.", targetClient)
	}

	if err := k.ConnectionKeeper.ConnOpenTry(
		ctx, msg.DesiredConnectionId, msg.CounterpartyChosenConnectionId, msg.Counterparty, msg.ClientId, targetClient,
		msg.CounterpartyVersions, msg.ProofInit, msg.ProofClient, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open try failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenTry,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, msg.DesiredConnectionId),
			sdk.NewAttribute(connectiontypes.AttributeKeyClientID, msg.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyClientID, msg.Counterparty.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyConnectionID, msg.Counterparty.ConnectionId),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, connectiontypes.AttributeValueCategory),
		),
	})

	return &connectiontypes.MsgConnectionOpenTryResponse{}, nil
}

func (k MsgServer) ConnectionOpenAck(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenAck) (*connectiontypes.MsgConnectionOpenAckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	targetClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "client in msg is not exported.ClientState. invalid client: %v", targetClient)
	}

	if err := k.ConnectionKeeper.ConnOpenAck(
		ctx, msg.ConnectionId, targetClient, msg.Version, msg.CounterpartyConnectionId,
		msg.ProofTry, msg.ProofClient, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open ack failed")
	}

	connectionEnd, _ := k.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenAck,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(connectiontypes.AttributeKeyClientID, connectionEnd.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyClientID, connectionEnd.Counterparty.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyConnectionID, connectionEnd.Counterparty.ConnectionId),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, connectiontypes.AttributeValueCategory),
		),
	})

	return &connectiontypes.MsgConnectionOpenAckResponse{}, nil
}

func (k MsgServer) ConnectionOpenConfirm(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenConfirm) (*connectiontypes.MsgConnectionOpenConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.ConnectionKeeper.ConnOpenConfirm(
		ctx, msg.ConnectionId, msg.ProofAck, msg.ProofHeight,
	); err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open confirm failed")
	}

	connectionEnd, _ := k.ConnectionKeeper.GetConnection(ctx, msg.ConnectionId)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenConfirm,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, msg.ConnectionId),
			sdk.NewAttribute(connectiontypes.AttributeKeyClientID, connectionEnd.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyClientID, connectionEnd.Counterparty.ClientId),
			sdk.NewAttribute(connectiontypes.AttributeKeyCounterpartyConnectionID, connectionEnd.Counterparty.ConnectionId),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, connectiontypes.AttributeValueCategory),
		),
	})

	return &connectiontypes.MsgConnectionOpenConfirmResponse{}, nil
}
