package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// VerifyClientConsensusState verifies a proof of the consensus state of the
// specified client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	ctx sdk.Context,
	clientState types.State,
	height uint64, // sequence
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	consensusState exported.ConsensusState,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.ConsensusStatePath(clientState.ID))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	bz, err := k.cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return false, err
	}

	return proof.VerifyMembership(root, path, bz), nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	connectionID string,
	connection connectiontypes.ConnectionEnd,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, connectiontypes.ConnectionPath(connectionID))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(connection)
	if err != nil {
		return false, err
	}

	return proof.VerifyMembership(root, path, bz), nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel channeltypes.Channel,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, channeltypes.ChannelPath(portID, channelID))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(channel)
	if err != nil {
		return false, err
	}

	return proof.VerifyMembership(root, path, bz), nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBz []byte,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, channeltypes.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyMembership(root, path, commitmentBz), nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, channeltypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyMembership(root, path, acknowledgement), nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, channeltypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyNonMembership(root, path), nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	clientState types.State,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) (bool, error) {
	if clientState.Frozen {
		return false, errors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, channeltypes.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return false, err
	}

	root, found := k.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, errors.ErrRootNotFound(k.codespace)
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	return proof.VerifyMembership(root, path, bz), nil
}
