package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// VerifyClientConsensusState verifies a proof of the consensus state of the
// specified client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	consensusState clientexported.ConsensusState,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return clienttypes.ErrClientNotFound
	}

	return clientState.VerifyClientConsensusState(
		k.cdc, height, connection.Counterparty.Prefix, proof, consensusState,
	)
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	ctx sdk.Context,
	height uint64,
	proof commitment.ProofI,
	connectionID string,
	connection types.ConnectionEnd,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyConnectionState(
		k.cdc, height, connection.Counterparty.Prefix, proof, connectionID, connection, consensusState,
	)
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel channelexported.ChannelI,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyChannelState(
		k.cdc, height, connection.GetCounterparty().GetPrefix(), proof,
		portID, channelID, channel, consensusState,
	)
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyPacketCommitment(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, commitmentBytes, consensusState,
	)
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyPacketAcknowledgement(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, acknowledgement, consensusState,
	)
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyPacketAcknowledgementAbsence(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		sequence, consensusState,
	)
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) error {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return clienttypes.ErrClientNotFound
	}

	consensusState, found := k.clientKeeper.GetClientConsensusState(
		ctx, connection.GetClientID(), height,
	)
	if !found {
		return clienttypes.ErrConsensusStateNotFound
	}

	return clientState.VerifyNextSequenceRecv(
		height, connection.GetCounterparty().GetPrefix(), proof, portID, channelID,
		nextSequenceRecv, consensusState,
	)
}
