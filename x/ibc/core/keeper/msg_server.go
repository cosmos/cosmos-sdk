package keeper

import (
	"context"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"
)

var _ clienttypes.MsgServer = Keeper{}
var _ connectiontypes.MsgServer = Keeper{}
var _ channeltypes.MsgServer = Keeper{}

// CreateClient defines a rpc handler method for MsgCreateClient.
func (k Keeper) CreateClient(goCtx context.Context, msg *clienttypes.MsgCreateClient) (*clienttypes.MsgCreateClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}

	consensusState, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return nil, err
	}

	clientID, err := k.ClientKeeper.CreateClient(ctx, clientState, consensusState)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			clienttypes.EventTypeCreateClient,
			sdk.NewAttribute(clienttypes.AttributeKeyClientID, clientID),
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

// UpdateClient defines a rpc handler method for MsgUpdateClient.
func (k Keeper) UpdateClient(goCtx context.Context, msg *clienttypes.MsgUpdateClient) (*clienttypes.MsgUpdateClientResponse, error) {
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

// UpgradeClient defines a rpc handler method for MsgUpgradeClient.
func (k Keeper) UpgradeClient(goCtx context.Context, msg *clienttypes.MsgUpgradeClient) (*clienttypes.MsgUpgradeClientResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	upgradedClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, err
	}
	upgradedConsState, err := clienttypes.UnpackConsensusState(msg.ConsensusState)
	if err != nil {
		return nil, err
	}

	if err = k.ClientKeeper.UpgradeClient(ctx, msg.ClientId, upgradedClient, upgradedConsState,
		msg.ProofUpgradeClient, msg.ProofUpgradeConsensusState); err != nil {
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

// SubmitMisbehaviour defines a rpc handler method for MsgSubmitMisbehaviour.
func (k Keeper) SubmitMisbehaviour(goCtx context.Context, msg *clienttypes.MsgSubmitMisbehaviour) (*clienttypes.MsgSubmitMisbehaviourResponse, error) {
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

// ConnectionOpenInit defines a rpc handler method for MsgConnectionOpenInit.
func (k Keeper) ConnectionOpenInit(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenInit) (*connectiontypes.MsgConnectionOpenInitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	connectionID, err := k.ConnectionKeeper.ConnOpenInit(ctx, msg.ClientId, msg.Counterparty, msg.Version, msg.DelayPeriod)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open init failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenInit,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, connectionID),
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

// ConnectionOpenTry defines a rpc handler method for MsgConnectionOpenTry.
func (k Keeper) ConnectionOpenTry(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenTry) (*connectiontypes.MsgConnectionOpenTryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	targetClient, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "client in msg is not exported.ClientState. invalid client: %v.", targetClient)
	}

	connectionID, err := k.ConnectionKeeper.ConnOpenTry(
		ctx, msg.PreviousConnectionId, msg.Counterparty, msg.DelayPeriod, msg.ClientId, targetClient,
		connectiontypes.ProtoVersionsToExported(msg.CounterpartyVersions), msg.ProofInit, msg.ProofClient, msg.ProofConsensus,
		msg.ProofHeight, msg.ConsensusHeight,
	)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "connection handshake open try failed")
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			connectiontypes.EventTypeConnectionOpenTry,
			sdk.NewAttribute(connectiontypes.AttributeKeyConnectionID, connectionID),
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

// ConnectionOpenAck defines a rpc handler method for MsgConnectionOpenAck.
func (k Keeper) ConnectionOpenAck(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenAck) (*connectiontypes.MsgConnectionOpenAckResponse, error) {
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

// ConnectionOpenConfirm defines a rpc handler method for MsgConnectionOpenConfirm.
func (k Keeper) ConnectionOpenConfirm(goCtx context.Context, msg *connectiontypes.MsgConnectionOpenConfirm) (*connectiontypes.MsgConnectionOpenConfirmResponse, error) {
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

// ChannelOpenInit defines a rpc handler method for MsgChannelOpenInit.
func (k Keeper) ChannelOpenInit(goCtx context.Context, msg *channeltypes.MsgChannelOpenInit) (*channeltypes.MsgChannelOpenInitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by port capability
	module, portCap, err := k.PortKeeper.LookupModuleByPort(ctx, msg.PortId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	_, channelID, cap, err := channel.HandleMsgChannelOpenInit(ctx, k.ChannelKeeper, portCap, msg)
	if err != nil {
		return nil, err
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	if err = cbs.OnChanOpenInit(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId, channelID, cap, msg.Channel.Counterparty, msg.Channel.Version); err != nil {
		return nil, sdkerrors.Wrap(err, "channel open init callback failed")
	}

	return &channeltypes.MsgChannelOpenInitResponse{}, nil
}

// ChannelOpenTry defines a rpc handler method for MsgChannelOpenTry.
func (k Keeper) ChannelOpenTry(goCtx context.Context, msg *channeltypes.MsgChannelOpenTry) (*channeltypes.MsgChannelOpenTryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Lookup module by port capability
	module, portCap, err := k.PortKeeper.LookupModuleByPort(ctx, msg.PortId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	_, channelID, cap, err := channel.HandleMsgChannelOpenTry(ctx, k.ChannelKeeper, portCap, msg)
	if err != nil {
		return nil, err
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	if err = cbs.OnChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId, channelID, cap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion); err != nil {
		return nil, sdkerrors.Wrap(err, "channel open try callback failed")
	}

	return &channeltypes.MsgChannelOpenTryResponse{}, nil
}

// ChannelOpenAck defines a rpc handler method for MsgChannelOpenAck.
func (k Keeper) ChannelOpenAck(goCtx context.Context, msg *channeltypes.MsgChannelOpenAck) (*channeltypes.MsgChannelOpenAckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	_, err = channel.HandleMsgChannelOpenAck(ctx, k.ChannelKeeper, cap, msg)
	if err != nil {
		return nil, err
	}

	if err = cbs.OnChanOpenAck(ctx, msg.PortId, msg.ChannelId, msg.CounterpartyVersion); err != nil {
		return nil, sdkerrors.Wrap(err, "channel open ack callback failed")
	}

	return &channeltypes.MsgChannelOpenAckResponse{}, nil
}

// ChannelOpenConfirm defines a rpc handler method for MsgChannelOpenConfirm.
func (k Keeper) ChannelOpenConfirm(goCtx context.Context, msg *channeltypes.MsgChannelOpenConfirm) (*channeltypes.MsgChannelOpenConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	_, err = channel.HandleMsgChannelOpenConfirm(ctx, k.ChannelKeeper, cap, msg)
	if err != nil {
		return nil, err
	}

	if err = cbs.OnChanOpenConfirm(ctx, msg.PortId, msg.ChannelId); err != nil {
		return nil, sdkerrors.Wrap(err, "channel open confirm callback failed")
	}

	return &channeltypes.MsgChannelOpenConfirmResponse{}, nil
}

// ChannelCloseInit defines a rpc handler method for MsgChannelCloseInit.
func (k Keeper) ChannelCloseInit(goCtx context.Context, msg *channeltypes.MsgChannelCloseInit) (*channeltypes.MsgChannelCloseInitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	if err = cbs.OnChanCloseInit(ctx, msg.PortId, msg.ChannelId); err != nil {
		return nil, sdkerrors.Wrap(err, "channel close init callback failed")
	}

	_, err = channel.HandleMsgChannelCloseInit(ctx, k.ChannelKeeper, cap, msg)
	if err != nil {
		return nil, err
	}

	return &channeltypes.MsgChannelCloseInitResponse{}, nil
}

// ChannelCloseConfirm defines a rpc handler method for MsgChannelCloseConfirm.
func (k Keeper) ChannelCloseConfirm(goCtx context.Context, msg *channeltypes.MsgChannelCloseConfirm) (*channeltypes.MsgChannelCloseConfirmResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	if err = cbs.OnChanCloseConfirm(ctx, msg.PortId, msg.ChannelId); err != nil {
		return nil, sdkerrors.Wrap(err, "channel close confirm callback failed")
	}

	_, err = channel.HandleMsgChannelCloseConfirm(ctx, k.ChannelKeeper, cap, msg)
	if err != nil {
		return nil, err
	}

	return &channeltypes.MsgChannelCloseConfirmResponse{}, nil
}

// RecvPacket defines a rpc handler method for MsgRecvPacket.
func (k Keeper) RecvPacket(goCtx context.Context, msg *channeltypes.MsgRecvPacket) (*channeltypes.MsgRecvPacketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.DestinationPort, msg.Packet.DestinationChannel)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	// Perform TAO verification
	if err := k.ChannelKeeper.RecvPacket(ctx, cap, msg.Packet, msg.ProofCommitment, msg.ProofHeight); err != nil {
		return nil, sdkerrors.Wrap(err, "receive packet verification failed")
	}

	// Perform application logic callback
	_, ack, err := cbs.OnRecvPacket(ctx, msg.Packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "receive packet callback failed")
	}

	// Set packet acknowledgement only if the acknowledgement is not nil.
	// NOTE: IBC applications modules may call the WriteAcknowledgement asynchronously if the
	// acknowledgement is nil.
	if ack != nil {
		if err := k.ChannelKeeper.WriteAcknowledgement(ctx, cap, msg.Packet, ack); err != nil {
			return nil, err
		}
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "ibc", msg.Type()},
			1,
			[]metrics.Label{
				telemetry.NewLabel("source-port", msg.Packet.SourcePort),
				telemetry.NewLabel("source-channel", msg.Packet.SourceChannel),
				telemetry.NewLabel("destination-port", msg.Packet.DestinationPort),
				telemetry.NewLabel("destination-channel", msg.Packet.DestinationChannel),
			},
		)
	}()

	return &channeltypes.MsgRecvPacketResponse{}, nil
}

// Timeout defines a rpc handler method for MsgTimeout.
func (k Keeper) Timeout(goCtx context.Context, msg *channeltypes.MsgTimeout) (*channeltypes.MsgTimeoutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	// Perform TAO verification
	if err := k.ChannelKeeper.TimeoutPacket(ctx, msg.Packet, msg.ProofUnreceived, msg.ProofHeight, msg.NextSequenceRecv); err != nil {
		return nil, sdkerrors.Wrap(err, "timeout packet verification failed")
	}

	// Perform application logic callback
	_, err = cbs.OnTimeoutPacket(ctx, msg.Packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "timeout packet callback failed")
	}

	// Delete packet commitment
	if err = k.ChannelKeeper.TimeoutExecuted(ctx, cap, msg.Packet); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", "timeout", "packet"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("source-port", msg.Packet.SourcePort),
				telemetry.NewLabel("source-channel", msg.Packet.SourceChannel),
				telemetry.NewLabel("destination-port", msg.Packet.DestinationPort),
				telemetry.NewLabel("destination-channel", msg.Packet.DestinationChannel),
				telemetry.NewLabel("timeout-type", "height"),
			},
		)
	}()

	return &channeltypes.MsgTimeoutResponse{}, nil
}

// TimeoutOnClose defines a rpc handler method for MsgTimeoutOnClose.
func (k Keeper) TimeoutOnClose(goCtx context.Context, msg *channeltypes.MsgTimeoutOnClose) (*channeltypes.MsgTimeoutOnCloseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	// Perform TAO verification
	if err := k.ChannelKeeper.TimeoutOnClose(ctx, cap, msg.Packet, msg.ProofUnreceived, msg.ProofClose, msg.ProofHeight, msg.NextSequenceRecv); err != nil {
		return nil, sdkerrors.Wrap(err, "timeout on close packet verification failed")
	}

	// Perform application logic callback
	// NOTE: MsgTimeout and MsgTimeoutOnClose use the same "OnTimeoutPacket"
	// application logic callback.
	_, err = cbs.OnTimeoutPacket(ctx, msg.Packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "timeout packet callback failed")
	}

	// Delete packet commitment
	if err = k.ChannelKeeper.TimeoutExecuted(ctx, cap, msg.Packet); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"ibc", "timeout", "packet"},
			1,
			[]metrics.Label{
				telemetry.NewLabel("source-port", msg.Packet.SourcePort),
				telemetry.NewLabel("source-channel", msg.Packet.SourceChannel),
				telemetry.NewLabel("destination-port", msg.Packet.DestinationPort),
				telemetry.NewLabel("destination-channel", msg.Packet.DestinationChannel),
				telemetry.NewLabel("timeout-type", "channel-closed"),
			},
		)
	}()

	return &channeltypes.MsgTimeoutOnCloseResponse{}, nil
}

// Acknowledgement defines a rpc handler method for MsgAcknowledgement.
func (k Keeper) Acknowledgement(goCtx context.Context, msg *channeltypes.MsgAcknowledgement) (*channeltypes.MsgAcknowledgementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Lookup module by channel capability
	module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
	}

	// Retrieve callbacks from router
	cbs, ok := k.Router.GetRoute(module)
	if !ok {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
	}

	// Perform TAO verification
	if err := k.ChannelKeeper.AcknowledgePacket(ctx, cap, msg.Packet, msg.Acknowledgement, msg.ProofAcked, msg.ProofHeight); err != nil {
		return nil, sdkerrors.Wrap(err, "acknowledge packet verification failed")
	}

	// Perform application logic callback
	_, err = cbs.OnAcknowledgementPacket(ctx, msg.Packet, msg.Acknowledgement)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "acknowledge packet callback failed")
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{"tx", "msg", "ibc", msg.Type()},
			1,
			[]metrics.Label{
				telemetry.NewLabel("source-port", msg.Packet.SourcePort),
				telemetry.NewLabel("source-channel", msg.Packet.SourceChannel),
				telemetry.NewLabel("destination-port", msg.Packet.DestinationPort),
				telemetry.NewLabel("destination-channel", msg.Packet.DestinationChannel),
			},
		)
	}()

	return &channeltypes.MsgAcknowledgementResponse{}, nil
}
