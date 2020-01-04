package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// CounterpartyHops returns the connection hops of the counterparty channel.
// The counterparty hops are stored in the inverse order as the channel's.
func (k Keeper) CounterpartyHops(ctx sdk.Context, ch types.Channel) ([]string, bool) {
	counterPartyHops := make([]string, len(ch.ConnectionHops))
	for i, hop := range ch.ConnectionHops {
		connection, found := k.connectionKeeper.GetConnection(ctx, hop)
		if !found {
			return []string{}, false
		}
		counterPartyHops[len(counterPartyHops)-1-i] = connection.Counterparty.ConnectionID
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
	counterparty types.Counterparty,
	version string,
) error {
	// TODO: abortTransactionUnless(validateChannelIdentifier(portIdentifier, channelIdentifier))
	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return sdkerrors.Wrap(types.ErrChannelExists, channelID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.State == connection.UNINITIALIZED {
		return sdkerrors.Wrap(
			connection.ErrInvalidConnectionState,
			"connection state cannot be UNINITIALIZED",
		)
	}

	/*
		// TODO: Maybe not right
		key := sdk.NewKVStoreKey(portID)

		if !k.portKeeper.Authenticate(key, portID) {
			return sdkerrors.Wrap(port.ErrInvalidPort, portID)
		}

	*/
	channel := types.NewChannel(types.INIT, order, counterparty, connectionHops, version)
	k.SetChannel(ctx, portID, channelID, channel)

	// TODO: generate channel capability key and set it to store
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return nil
}

// ChanOpenTry is called by a module to accept the first step of a channel opening
// handshake initiated by a module on another chain.
func (k Keeper) ChanOpenTry(
	ctx sdk.Context,
	order types.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty types.Counterparty,
	version,
	counterpartyVersion string,
	proofInit commitment.ProofI,
	proofHeight uint64,
) error {
	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return sdkerrors.Wrap(types.ErrChannelExists, channelID)
	}

	// TODO: Maybe not right
	key := sdk.NewKVStoreKey(portID)

	if !k.portKeeper.Authenticate(key, portID) {
		return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
		)
	}

	// NOTE: this step has been switched with the one below to reverse the connection
	// hops
	channel := types.NewChannel(types.TRYOPEN, order, counterparty, connectionHops, version)

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// expectedCounterpaty is the counterparty of the counterparty's channel end
	// (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.INIT, channel.Ordering, expectedCounterparty,
		counterpartyHops, channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return err
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proofInit,
		types.ChannelPath(counterparty.PortID, counterparty.ChannelID),
		bz,
	) {
		return sdkerrors.Wrap(types.ErrInvalidCounterparty, "channel membership verification failed")
	}

	k.SetChannel(ctx, portID, channelID, channel)

	// TODO: generate channel capability key and set it to store
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return nil
}

// ChanOpenAck is called by the handshake-originating module to acknowledge the
// acceptance of the initial request by the counterparty module on the other chain.
func (k Keeper) ChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyVersion string,
	proofTry commitment.ProofI,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State != types.INIT {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not INIT (got %s)", channel.State.String(),
		)
	}

	// TODO: Maybe not right
	key := sdk.NewKVStoreKey(portID)

	if !k.portKeeper.Authenticate(key, portID) {
		return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// counterparty of the counterparty channel end (i.e self)
	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		types.TRYOPEN, channel.Ordering, counterparty,
		counterpartyHops, channel.Version,
	)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return err
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proofTry,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return sdkerrors.Wrap(
			types.ErrInvalidCounterparty, "channel membership verification failed",
		)
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
	proofAck commitment.ProofI,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State != types.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not OPENTRY (got %s)", channel.State.String(),
		)
	}

	// TODO: Maybe not right
	key := sdk.NewKVStoreKey(portID)

	if !k.portKeeper.Authenticate(key, portID) {
		return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
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

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return err
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proofAck,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return sdkerrors.Wrap(types.ErrInvalidCounterparty, "channel membership verification failed")
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
func (k Keeper) ChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// TODO: Maybe not right
	key := sdk.NewKVStoreKey(portID)

	if !k.portKeeper.Authenticate(key, portID) {
		return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State == types.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
		)
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
	proofInit commitment.ProofI,
	proofHeight uint64,
) error {
	// TODO: Maybe not right
	key := sdk.NewKVStoreKey(portID)

	if !k.portKeeper.Authenticate(key, portID) {
		return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State == types.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.State != connection.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.State.String(),
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

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(expectedChannel)
	if err != nil {
		return errors.New("failed to marshal expected channel")
	}

	if !k.connectionKeeper.VerifyMembership(
		ctx, connectionEnd, proofHeight, proofInit,
		types.ChannelPath(channel.Counterparty.PortID, channel.Counterparty.ChannelID),
		bz,
	) {
		return sdkerrors.Wrap(types.ErrInvalidCounterparty, "channel membership verification failed")
	}

	channel.State = types.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}
