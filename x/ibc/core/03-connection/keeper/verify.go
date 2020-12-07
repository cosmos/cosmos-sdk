package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// VerifyClientState verifies a proof of a client state of the running machine
// stored on the target machine
func (k Keeper) VerifyClientState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	clientState exported.ClientState,
) error {
	clientID := connection.GetClientID()
	targetClient, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	if err := targetClient.VerifyClientState(
		k.clientKeeper.ClientStore(ctx, clientID), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), connection.GetCounterparty().GetClientID(), proof, clientState); err != nil {
		return sdkerrors.Wrapf(err, "failed client state verification for target client: %s", connection.GetClientID())
	}

	return nil
}

// VerifyClientConsensusState verifies a proof of the consensus state of the
// specified client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	consensusHeight exported.Height,
	proof []byte,
	consensusState exported.ConsensusState,
) error {
	clientID := connection.GetClientID()
	clientState, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	if err := clientState.VerifyClientConsensusState(
		k.clientKeeper.ClientStore(ctx, clientID), k.cdc, height,
		connection.GetCounterparty().GetClientID(), consensusHeight, connection.GetCounterparty().GetPrefix(), proof, consensusState,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed consensus state verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	connectionID string,
	connectionEnd exported.ConnectionI, // opposite connection
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyConnectionState(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, connectionID, connectionEnd,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed connection state verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	portID,
	channelID string,
	channel exported.ChannelI,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyChannelState(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof,
		portID, channelID, channel,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed channel state verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyPacketCommitment(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		uint64(ctx.BlockTime().UnixNano()), connection.GetDelayPeriod(),
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, commitmentBytes,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed packet commitment verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyPacketAcknowledgement(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		uint64(ctx.BlockTime().UnixNano()), connection.GetDelayPeriod(),
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, acknowledgement,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed packet acknowledgement verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyPacketReceiptAbsence verifies a proof of the absence of an
// incoming packet receipt at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketReceiptAbsence(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyPacketReceiptAbsence(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		uint64(ctx.BlockTime().UnixNano()), connection.GetDelayPeriod(),
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed packet receipt absence verification for client (%s)", connection.GetClientID())
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height exported.Height,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	if err := clientState.VerifyNextSequenceRecv(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		uint64(ctx.BlockTime().UnixNano()), connection.GetDelayPeriod(),
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		nextSequenceRecv,
	); err != nil {
		return sdkerrors.Wrapf(err, "failed next sequence receive verification for client (%s)", connection.GetClientID())
	}

	return nil
}
