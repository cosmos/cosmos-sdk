package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienterrors "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
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
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyClientConsensusState(
		ctx, clientState, height, connection.Counterparty.Prefix, proof, consensusState,
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
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyConnectionState(
		ctx, clientState, height, connection.Counterparty.Prefix, proof, connectionID, connection,
	)
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel channeltypes.Channel,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyChannelState(
		ctx, clientState, height, prefix, proof, portID, channelID, channel,
	)
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBz []byte,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyPacketCommitment(
		ctx, clientState, height, prefix, proof, portID, channelID, sequence, commitmentBz,
	)
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyPacketAcknowledgement(
		ctx, clientState, height, prefix, proof, portID, channelID, sequence, acknowledgement,
	)
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyPacketAcknowledgementAbsence(
		ctx, clientState, height, prefix, proof, portID, channelID, sequence,
	)
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connection types.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connection.ClientID)
	}

	return k.clientKeeper.VerifyNextSequenceRecv(
		ctx, clientState, height, prefix, proof, portID, channelID, nextSequenceRecv,
	)
}
