package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
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
		counterPartyHops[len(counterPartyHops)-1-i] = connection.GetCounterparty().GetConnectionID()
	}
	return counterPartyHops, true
}

// ChanOpenInit is called by a module to initiate a channel opening handshake with
// a module on another chain.
func (k Keeper) ChanOpenInit(
	ctx sdk.Context,
	order ibctypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty types.Counterparty,
	version string,
) error {
	// channel identifier and connection hop length checked on msg.ValidateBasic()

	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return sdkerrors.Wrap(types.ErrChannelExists, channelID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.GetState() == ibctypes.UNINITIALIZED {
		return sdkerrors.Wrap(
			connection.ErrInvalidConnectionState,
			"connection state cannot be UNINITIALIZED",
		)
	}

	channel := types.NewChannel(ibctypes.INIT, order, counterparty, connectionHops, version)
	k.SetChannel(ctx, portID, channelID, channel)

	// TODO: blocked by #5542
	// key := ""
	// k.SetChannelCapability(ctx, portID, channelID, key)
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return nil
}

// ChanOpenTry is called by a module to accept the first step of a channel opening
// handshake initiated by a module on another chain.
func (k Keeper) ChanOpenTry(
	ctx sdk.Context,
	order ibctypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty types.Counterparty,
	version,
	counterpartyVersion string,
	proofInit commitmentexported.Proof,
	proofHeight uint64,
) error {
	// channel identifier and connection hop length checked on msg.ValidateBasic()

	previousChannel, found := k.GetChannel(ctx, portID, channelID)
	if found && !(previousChannel.GetState() == ibctypes.INIT &&
		previousChannel.GetOrdering() == order &&
		previousChannel.GetCounterparty().GetPortID() == counterparty.PortID &&
		previousChannel.GetCounterparty().GetChannelID() == counterparty.ChannelID &&
		previousChannel.GetConnectionHops()[0] == connectionHops[0] &&
		previousChannel.GetVersion() == version) {
		sdkerrors.Wrap(types.ErrInvalidChannel, "cannot relay connection attempt")
	}

	// TODO: blocked by #5542
	// key := sdk.NewKVStoreKey(portID)
	// if !k.portKeeper.Authenticate(key, portID) {
	// 	return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	// }

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.GetState() != ibctypes.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	// NOTE: this step has been switched with the one below to reverse the connection
	// hops
	channel := types.NewChannel(ibctypes.TRYOPEN, order, counterparty, connectionHops, version)

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// expectedCounterpaty is the counterparty of the counterparty's channel end
	// (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		ibctypes.INIT, channel.GetOrdering(), expectedCounterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		counterparty.PortID, counterparty.ChannelID, expectedChannel,
	); err != nil {
		return err
	}

	k.SetChannel(ctx, portID, channelID, channel)

	// TODO: blocked by #5542
	// key := ""
	// k.SetChannelCapability(ctx, portID, channelID, key)
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
	proofTry commitmentexported.Proof,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if !(channel.GetState() == ibctypes.INIT || channel.GetState() == ibctypes.TRYOPEN) {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state should be INIT or TRYOPEN (got %s)", channel.GetState().String(),
		)
	}

	// TODO: blocked by #5542
	// key := sdk.NewKVStoreKey(portID)
	// if !k.portKeeper.Authenticate(key, portID) {
	// 	return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	// }

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.GetConnectionHops()[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.GetConnectionHops()[0])
	}

	if connectionEnd.GetState() != ibctypes.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// counterparty of the counterparty channel end (i.e self)
	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		ibctypes.TRYOPEN, channel.GetOrdering(), counterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofTry,
		channel.GetCounterparty().GetPortID(),
		channel.GetCounterparty().GetChannelID(),
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = ibctypes.OPEN
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
	proofAck commitmentexported.Proof,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.GetState() != ibctypes.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not TRYOPEN (got %s)", channel.GetState().String(),
		)
	}

	// TODO: blocked by #5542
	// capkey, found := k.GetChannelCapability(ctx, portID, channelID)
	// if !found {
	// 	return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, channelID)
	// }

	// key := sdk.NewKVStoreKey(capkey)
	// if !k.portKeeper.Authenticate(key, portID) {
	// 	return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	// }

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.GetConnectionHops()[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.GetConnectionHops()[0])
	}

	if connectionEnd.GetState() != ibctypes.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		ibctypes.OPEN, channel.GetOrdering(), counterparty,
		counterpartyHops, channel.GetVersion(),
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofAck,
		channel.GetCounterparty().GetPortID(), channel.GetCounterparty().GetPortID(),
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = ibctypes.OPEN
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
	// TODO: blocked by #5542
	// capkey, found := k.GetChannelCapability(ctx, portID, channelID)
	// if !found {
	// 	return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, channelID)
	// }

	// key := sdk.NewKVStoreKey(capkey)
	// if !k.portKeeper.Authenticate(key, portID) {
	// 	return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	// }

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.GetState() == ibctypes.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.GetConnectionHops()[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.GetConnectionHops()[0])
	}

	if connectionEnd.GetState() != ibctypes.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	channel.State = ibctypes.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)
	k.Logger(ctx).Info("channel close initialized: portID (%s), channelID (%s)", portID, channelID)
	return nil
}

// ChanCloseConfirm is called by the counterparty module to close their end of the
// channel, since the other end has been closed.
func (k Keeper) ChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	proofInit commitmentexported.Proof,
	proofHeight uint64,
) error {
	// TODO: blocked by #5542
	// capkey, found := k.GetChannelCapability(ctx, portID, channelID)
	// if !found {
	// 	return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, channelID)
	// }

	// key := sdk.NewKVStoreKey(capkey)
	// if !k.portKeeper.Authenticate(key, portID) {
	// 	return sdkerrors.Wrap(port.ErrInvalidPort, portID)
	// }

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.GetState() == ibctypes.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.GetConnectionHops()[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.GetConnectionHops()[0])
	}

	if connectionEnd.GetState() != ibctypes.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		ibctypes.CLOSED, channel.GetOrdering(), counterparty,
		counterpartyHops, channel.GetVersion(),
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		channel.GetCounterparty().GetPortID(), channel.GetCounterparty().GetChannelID(),
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = ibctypes.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}
