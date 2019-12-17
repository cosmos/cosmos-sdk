package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienterrors "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	ctx sdk.Context,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel types.Channel,
	connectionEnd connection.ConnectionEnd,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connectionEnd.ClientID)
	}

	if clientState.Frozen {
		return false, clienterrors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.ChannelPath(portID, channelID))
	if err != nil {
		return false, err
	}

	root, found := k.clientKeeper.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, clienterrors.ErrRootNotFound(k.codespace)
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
	connectionEnd connection.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBz []byte,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connectionEnd.ClientID)
	}

	if clientState.Frozen {
		return false, clienterrors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.clientKeeper.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, clienterrors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyMembership(root, path, commitmentBz), nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	ctx sdk.Context,
	connectionEnd connection.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connectionEnd.ClientID)
	}

	if clientState.Frozen {
		return false, clienterrors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.clientKeeper.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, clienterrors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyMembership(root, path, acknowledgement), nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	ctx sdk.Context,
	connectionEnd connection.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connectionEnd.ClientID)
	}

	if clientState.Frozen {
		return false, clienterrors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return false, err
	}

	root, found := k.clientKeeper.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, clienterrors.ErrRootNotFound(k.codespace)
	}

	return proof.VerifyNonMembership(root, path), nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	ctx sdk.Context,
	connectionEnd connection.ConnectionEnd,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
) (bool, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connectionEnd.ClientID)
	if !found {
		return false, clienterrors.ErrClientNotFound(k.codespace, connectionEnd.ClientID)
	}

	if clientState.Frozen {
		return false, clienterrors.ErrClientFrozen(k.codespace, clientState.ID)
	}

	path, err := commitment.ApplyPrefix(prefix, types.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return false, err
	}

	root, found := k.clientKeeper.GetVerifiedRoot(ctx, clientState.ID, height)
	if !found {
		return false, clienterrors.ErrRootNotFound(k.codespace)
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	return proof.VerifyMembership(root, path, bz), nil
}
