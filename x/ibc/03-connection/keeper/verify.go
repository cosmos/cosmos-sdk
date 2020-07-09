package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
)

// VerifyClientConsensusState verifies a proof of the consensus state of the
// specified client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	consensusHeight uint64,
	proof []byte,
	consensusState clientexported.ConsensusState,
) error {
	clientID := connection.GetClientID()
	clientState, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	targetConsState, found := k.clientKeeper.GetClientConsensusState(ctx, clientID, height)
	if !found {
		return sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "clientID: %s with height: %d", clientID, height)
	}

	if err := clientState.VerifyClientConsensusState(
		k.clientKeeper.ClientStore(ctx, clientID), k.cdc, k.aminoCdc, targetConsState.GetRoot(), height,
		connection.GetCounterparty().GetClientID(), consensusHeight, connection.GetCounterparty().GetPrefix(), proof, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed consensus state verification")
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof []byte,
	connectionID string,
	connectionEnd exported.ConnectionI, // opposite connection
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyConnectionState(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, connectionID, connectionEnd, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed connection state verification")
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof []byte,
	portID,
	channelID string,
	channel channelexported.ChannelI,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyChannelState(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof,
		portID, channelID, channel, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed channel state verification")
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
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

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyPacketCommitment(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, commitmentBytes, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed packet commitment verification")
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
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

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyPacketAcknowledgement(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, acknowledgement, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed packet acknowledgement verification")
	}

	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof []byte,
	portID,
	channelID string,
	sequence uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyPacketAcknowledgementAbsence(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed packet acknowledgement absence verification")
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connection exported.ConnectionI,
	height uint64,
	proof []byte,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return sdkerrors.Wrap(clienttypes.ErrClientNotFound, connection.GetClientID())
	}

	// TODO: move to specific clients; blocked by #5502
	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return sdkerrors.Wrapf(
			clienttypes.ErrConsensusStateNotFound,
			"clientID (%s), height (%d)", connection.GetClientID(), height,
		)
	}

	if err := clientState.VerifyNextSequenceRecv(
		k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height,
		connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		nextSequenceRecv, consensusState,
	); err != nil {
		return sdkerrors.Wrap(err, "failed next sequence receive verification")
	}

	return nil
}
