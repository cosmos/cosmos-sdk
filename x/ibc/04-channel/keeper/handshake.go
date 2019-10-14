package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ics03types "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ChanOpenInit is called by a module to initiate a channel opening handshake with
// a module on another chain.
func (k Keeper) ChanOpenInit(
	ctx sdk.Context,
	order types.ChannelOrder,
	connectionHops []string,
	portID,
	channelID string,
	counterparty types.Counterparty,
	version string,
) (string, error) {
	// TODO: abortTransactionUnless(validateChannelIdentifier(portIdentifier, channelIdentifier))
	if len(connectionHops) != 1 {
		return "", types.ErrInvalidConnectionHops(k.codespace)
	}

	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return "", types.ErrChannelExists(k.codespace)
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return "", ics03types.ErrConnectionNotFound(k.codespace)
	}

	// TODO: inconsistency on ICS03 (`none`) and ICS04 (`CLOSED`)
	if connection.State == ics03types.NONE {
		return "", errors.New("connection is closed")
	}

	// TODO: Blocked - ICS05 Not implemented yet
	// port, found := k.portKeeper.GetPort(ctx, portID)
	// if !found {
	// 	return errors.New("port not found") // TODO: ics05 sdk.Error
	// }

	// if !k.portKeeper.AuthenticatePort(port.ID()) {
	// 	return errors.New("port is not valid") // TODO: ics05 sdk.Error
	// }

	channel := types.NewChannel(types.INIT, order, counterparty, connectionHops, version)
	k.SetChannel(ctx, portID, channelID, channel)

	key := "" // TODO: generate key
	k.SetChannelCapability(ctx, portID, channelID, key)
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return key, nil
}

// ChanOpenTry is called by a module to accept the first step of a channel opening
// handshake initiated by a module on another chain.
func (k Keeper) ChanOpenTry(
	ctx sdk.Context,
	order types.ChannelOrder,
	connectionHops []string,
	portID,
	channelID string,
	counterparty types.Counterparty,
	version,
	counterpartyVersion string,
	proofInit ics23.Proof,
	proofHeight uint64,
) (string, error) {

	if len(connectionHops) != 1 {
		return "", types.ErrInvalidConnectionHops(k.codespace)
	}

	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return "", types.ErrChannelExists(k.codespace)
	}

	// TODO: Blocked - ICS05 Not implemented yet
	// port, found := k.portKeeper.GetPort(ctx, portID)
	// if !found {
	// 	return errors.New("port not found") // TODO: ics05 sdk.Error
	// }

	// if !k.portKeeper.AuthenticatePort(port.ID()) {
	// 	return errors.New("port is not valid") // TODO: ics05 sdk.Error
	// }

	connection, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return "", ics03types.ErrConnectionNotFound(k.codespace)
	}

	if connection.State != ics03types.OPEN {
		return "", errors.New("connection is not open")
	}

	// NOTE: this step has been switched with the one below to reverse the connection
	// hops
	channel := types.NewChannel(types.OPENTRY, order, counterparty, connectionHops, version)

	// expectedCounterpaty is the counterparty of the counterparty's channel end
	// (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.INIT, channel.Ordering, expectedCounterparty,
		channel.CounterpartyHops(), channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return "", errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proofInit,
		types.ChannelPath(counterparty.PortID, counterparty.ChannelID),
		bz,
	) {
		return "", types.ErrInvalidCounterpartyChannel(k.codespace)
	}

	k.SetChannel(ctx, portID, channelID, channel)

	key := "" // TODO: generate key
	k.SetChannelCapability(ctx, portID, channelID, key)
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return key, nil
}

// ChanOpenAck is called by the handshake-originating module to acknowledge the
// acceptance of the initial request by the counterparty module on the other chain.
func (k Keeper) ChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyVersion string,
	proofTry ics23.Proof,
	proofHeight uint64,
) error {

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound(k.codespace)
	}

	if channel.State != types.INIT {
		return errors.New("invalid channel state") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, portID, channelID)
	if !found {
		return types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return ics03types.ErrConnectionNotFound(k.codespace)
	}

	if connection.State != ics03types.OPEN {
		return errors.New("connection is not open")
	}

	// counterparty of the counterparty channel end (i.e self)
	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.INIT, channel.Ordering, counterparty,
		channel.CounterpartyHops(), channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proofTry,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return types.ErrInvalidCounterpartyChannel(k.codespace)
	}

	channel.State = types.OPEN
	channel.Version = counterpartyVersion
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// ChanOpenConfirm is called by the counterparty module to close their end of the
//  channel, since the other end has been closed.
func (k Keeper) ChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	proofAck ics23.Proof,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound(k.codespace)
	}

	if channel.State != types.OPENTRY {
		return errors.New("invalid channel state") // TODO: sdk.Error
	}

	_, found = k.GetChannelCapability(ctx, portID, channelID)
	if !found {
		return types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return ics03types.ErrConnectionNotFound(k.codespace)
	}

	if connection.State != ics03types.OPEN {
		return errors.New("connection is not open")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.OPEN, channel.Ordering, counterparty,
		channel.CounterpartyHops(), channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proofAck,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return types.ErrInvalidCounterpartyChannel(k.codespace)
	}

	channel.State = types.OPEN
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// Closing Handshake
//
// This section defines the set of functions required to close a channel handshake
// as defined in https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#closing-handshake

// ChanCloseInit is called by either module to close their end of the channel. Once
// closed, channels cannot be reopened.
func (k Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	_, found := k.GetChannelCapability(ctx, portID, channelID)
	if !found {
		return types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound(k.codespace)
	}

	if channel.State == types.CLOSED {
		return errors.New("channel already closed") // TODO: sdk.Error
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return ics03types.ErrConnectionNotFound(k.codespace)
	}

	if connection.State != ics03types.OPEN {
		return errors.New("connection is not open")
	}

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
	proofInit ics23.Proof,
	proofHeight uint64,
) error {
	_, found := k.GetChannelCapability(ctx, portID, channelID)
	if !found {
		return types.ErrChannelCapabilityNotFound(k.codespace)
	}

	// if !AuthenticateCapabilityKey(capabilityKey) {
	//  return errors.New("invalid capability key") // TODO: sdk.Error
	// }

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return types.ErrChannelNotFound(k.codespace)
	}

	if channel.State == types.CLOSED {
		return errors.New("channel already closed") // TODO: sdk.Error
	}

	connection, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return ics03types.ErrConnectionNotFound(k.codespace)
	}

	if connection.State != ics03types.OPEN {
		return errors.New("connection is not open")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.CLOSED, channel.Ordering, counterparty,
		channel.CounterpartyHops(), channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connection, proofHeight, proofInit,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return types.ErrInvalidCounterpartyChannel(k.codespace)
	}

	channel.State = types.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}
