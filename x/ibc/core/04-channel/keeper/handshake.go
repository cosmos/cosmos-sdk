package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CounterpartyHops returns the connection hops of the counterparty channel.
// The counterparty hops are stored in the inverse order as the channel's.
func (k Keeper) CounterpartyHops(ctx sdk.Context, ch types.Channel) ([]string, bool) {
	counterPartyHops := make([]string, len(ch.ConnectionHops))

	for i, hop := range ch.ConnectionHops {
		conn, found := k.connectionKeeper.GetConnection(ctx, hop)
		if !found {
			return []string{}, false
		}

		counterPartyHops[len(counterPartyHops)-1-i] = conn.GetCounterparty().GetConnectionID()
	}

	return counterPartyHops, true
}

// ChanOpenInit is called by a module to initiate a channel opening handshake with
// a module on another chain.
func (k Keeper) ChanOpenInit(
	ctx sdk.Context,
	order types.Order,
	connectionHops []string,
	portID,
	channelID string,
	portCap *capabilitytypes.Capability,
	counterparty types.Counterparty,
	version string,
) (*capabilitytypes.Capability, error) {
	// channel identifier and connection hop length checked on msg.ValidateBasic()
	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrChannelExists, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionHops[0])
	}

	if len(connectionEnd.GetVersions()) != 1 {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidVersion,
			"single version must be negotiated on connection before opening channel, got: %v",
			connectionEnd.GetVersions(),
		)
	}

	if !connectiontypes.VerifySupportedFeature(connectionEnd.GetVersions()[0], order.String()) {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidVersion,
			"connection version %s does not support channel ordering: %s",
			connectionEnd.GetVersions()[0], order.String(),
		)
	}

	if !k.portKeeper.Authenticate(ctx, portCap, portID) {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidPort, "caller does not own port capability for port ID %s", portID)
	}

	channel := types.NewChannel(types.INIT, order, counterparty, connectionHops, version)
	k.SetChannel(ctx, portID, channelID, channel)

	capKey, err := k.scopedKeeper.NewCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "could not create channel capability for port ID %s and channel ID %s", portID, channelID)
	}

	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)
	k.SetNextSequenceAck(ctx, portID, channelID, 1)

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: NONE -> INIT", portID, channelID))
	return capKey, nil
}

// ChanOpenTry is called by a module to accept the first step of a channel opening
// handshake initiated by a module on another chain.
func (k Keeper) ChanOpenTry(
	ctx sdk.Context,
	order types.Order,
	connectionHops []string,
	portID,
	desiredChannelID,
	counterpartyChosenChannelID string,
	portCap *capabilitytypes.Capability,
	counterparty types.Counterparty,
	version,
	counterpartyVersion string,
	proofInit []byte,
	proofHeight exported.Height,
) (*capabilitytypes.Capability, error) {
	// channel identifier and connection hop length checked on msg.ValidateBasic()
	previousChannel, found := k.GetChannel(ctx, portID, desiredChannelID)
	if found && !(previousChannel.State == types.INIT &&
		previousChannel.Ordering == order &&
		previousChannel.Counterparty.PortId == counterparty.PortId &&
		previousChannel.Counterparty.ChannelId == counterparty.ChannelId &&
		previousChannel.ConnectionHops[0] == connectionHops[0] &&
		previousChannel.Version == version) {
		return nil, sdkerrors.Wrap(types.ErrInvalidChannel, "cannot relay connection attempt")
	}

	if !k.portKeeper.Authenticate(ctx, portCap, portID) {
		return nil, sdkerrors.Wrapf(porttypes.ErrInvalidPort, "caller does not own port capability for port ID %s", portID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
		)
	}

	if len(connectionEnd.GetVersions()) != 1 {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidVersion,
			"single version must be negotiated on connection before opening channel, got: %v",
			connectionEnd.GetVersions(),
		)
	}

	if !connectiontypes.VerifySupportedFeature(connectionEnd.GetVersions()[0], order.String()) {
		return nil, sdkerrors.Wrapf(
			connectiontypes.ErrInvalidVersion,
			"connection version %s does not support channel ordering: %s",
			connectionEnd.GetVersions()[0], order.String(),
		)
	}

	// If the channel id chosen for this channel end by the counterparty is empty then
	// flexible channel identifier selection is allowed by using the desired channel id.
	// Otherwise the desiredChannelID must match the counterpartyChosenChannelID.
	if counterpartyChosenChannelID != "" && counterpartyChosenChannelID != desiredChannelID {
		return nil, sdkerrors.Wrapf(
			types.ErrInvalidChannelIdentifier,
			"counterparty chosen channel ID (%s) must be empty or equal to the desired channel ID (%s)", counterpartyChosenChannelID, desiredChannelID,
		)
	}

	// NOTE: this step has been switched with the one below to reverse the connection
	// hops
	channel := types.NewChannel(types.TRYOPEN, order, counterparty, connectionHops, version)

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// expectedCounterpaty is the counterparty of the counterparty's channel end
	// (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, counterpartyChosenChannelID)
	expectedChannel := types.NewChannel(
		types.INIT, channel.Ordering, expectedCounterparty,
		counterpartyHops, counterpartyVersion,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		counterparty.PortId, counterparty.ChannelId, expectedChannel,
	); err != nil {
		return nil, err
	}

	k.SetChannel(ctx, portID, desiredChannelID, channel)

	capKey, err := k.scopedKeeper.NewCapability(ctx, host.ChannelCapabilityPath(portID, desiredChannelID))
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "could not create channel capability for port ID %s and channel ID %s", portID, desiredChannelID)
	}

	k.SetNextSequenceSend(ctx, portID, desiredChannelID, 1)
	k.SetNextSequenceRecv(ctx, portID, desiredChannelID, 1)
	k.SetNextSequenceAck(ctx, portID, desiredChannelID, 1)

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: NONE -> TRYOPEN", portID, desiredChannelID))
	return capKey, nil
}

// ChanOpenAck is called by the handshake-originating module to acknowledge the
// acceptance of the initial request by the counterparty module on the other chain.
func (k Keeper) ChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterpartyVersion,
	counterpartyChannelID string,
	proofTry []byte,
	proofHeight exported.Height,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	if !(channel.State == types.INIT || channel.State == types.TRYOPEN) {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state should be INIT or TRYOPEN (got %s)", channel.State.String(),
		)
	}

	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrapf(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel, port ID (%s) channel ID (%s)", portID, channelID)
	}

	// If the previously set channel end allowed for the counterparty to select its own
	// channel identifier then we use the counterpartyChannelID. Otherwise the
	// counterpartyChannelID must match the previously set counterparty channel ID.
	if channel.Counterparty.ChannelId != "" && counterpartyChannelID != channel.Counterparty.ChannelId {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelIdentifier,
			"counterparty channel identifier (%s) must be equal to stored channel ID for counterparty (%s)", counterpartyChannelID, channel.Counterparty.ChannelId,
		)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
		return sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// counterparty of the counterparty channel end (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.TRYOPEN, channel.Ordering, expectedCounterparty,
		counterpartyHops, counterpartyVersion,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofTry,
		channel.Counterparty.PortId, counterpartyChannelID,
		expectedChannel,
	); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: %s -> OPEN", portID, channelID, channel.State))

	channel.State = types.OPEN
	channel.Version = counterpartyVersion
	channel.Counterparty.ChannelId = counterpartyChannelID
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// ChanOpenConfirm is called by the counterparty module to close their end of the
//  channel, since the other end has been closed.
func (k Keeper) ChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	proofAck []byte,
	proofHeight exported.Height,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	if channel.State != types.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not TRYOPEN (got %s)", channel.State.String(),
		)
	}

	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrapf(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel, port ID (%s) channel ID (%s)", portID, channelID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
		return sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.OPEN, channel.Ordering, counterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofAck,
		channel.Counterparty.PortId, channel.Counterparty.ChannelId,
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = types.OPEN
	k.SetChannel(ctx, portID, channelID, channel)

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: TRYOPEN -> OPEN", portID, channelID))
	return nil
}

// Closing Handshake
//
// This section defines the set of functions required to close a channel handshake
// as defined in https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#closing-handshake

// ChanCloseInit is called by either module to close their end of the channel. Once
// closed, channels cannot be reopened.
func (k Keeper) ChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
) error {
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrapf(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel, port ID (%s) channel ID (%s)", portID, channelID)
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	if channel.State == types.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
		return sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
		)
	}

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: %s -> CLOSED", portID, channelID, channel.State))

	channel.State = types.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// ChanCloseConfirm is called by the counterparty module to close their end of the
// channel, since the other end has been closed.
func (k Keeper) ChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	proofInit []byte,
	proofHeight exported.Height,
) error {
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel, port ID (%s) channel ID (%s)")
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrapf(types.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}

	if channel.State == types.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connectiontypes.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != int32(connectiontypes.OPEN) {
		return sdkerrors.Wrapf(
			connectiontypes.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectiontypes.State(connectionEnd.GetState()).String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.CLOSED, channel.Ordering, counterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		channel.Counterparty.PortId, channel.Counterparty.ChannelId,
		expectedChannel,
	); err != nil {
		return err
	}

	k.Logger(ctx).Info(fmt.Sprintf("channel (port-id: %s, channel-id: %s) state updated: %s -> CLOSED", portID, channelID, channel.State))

	channel.State = types.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}
