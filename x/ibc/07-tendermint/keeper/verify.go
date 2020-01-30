package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// VerifyClientConsensusState verifies a proof of the consensus state of the
// Tendermint client stored on the target machine.
func (k Keeper) VerifyClientConsensusState(
	cdc *codec.Codec,
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ConsensusStatePath(cs.GetID(), height))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	bz, err := cdc.MarshalBinaryBare(consensusState)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedClientConsensusStateVerification, err.Error())
	}

	return nil
}

// VerifyConnectionState verifies a proof of the connection state of the
// specified connection end stored on the target machine.
func (k Keeper) VerifyConnectionState(
	cdc *codec.Codec,
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	connectionID string,
	connectionEnd connectionexported.ConnectionI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ConnectionPath(connectionID))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	bz, err := cdc.MarshalBinaryBare(connectionEnd)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedConnectionStateVerification, err.Error())
	}

	return nil
}

// VerifyChannelState verifies a proof of the channel state of the specified
// channel end, under the specified port, stored on the target machine.
func (k Keeper) VerifyChannelState(
	cdc *codec.Codec,
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	channel channelexported.ChannelI,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.ChannelPath(portID, channelID))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	bz, err := cdc.MarshalBinaryBare(channel)
	if err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedChannelStateVerification, err.Error())
	}

	return nil
}

// VerifyPacketCommitment verifies a proof of an outgoing packet commitment at
// the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketCommitment(
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	commitmentBytes []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketCommitmentPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, commitmentBytes); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketCommitmentVerification, err.Error())
	}

	return nil
}

// VerifyPacketAcknowledgement verifies a proof of an incoming packet
// acknowledgement at the specified port, specified channel, and specified sequence.
func (k Keeper) VerifyPacketAcknowledgement(
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	acknowledgement []byte,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, acknowledgement); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckVerification, err.Error())
	}

	return nil
}

// VerifyPacketAcknowledgementAbsence verifies a proof of the absence of an
// incoming packet acknowledgement at the specified port, specified channel, and
// specified sequence.
func (k Keeper) VerifyPacketAcknowledgementAbsence(
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	sequence uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.PacketAcknowledgementPath(portID, channelID, sequence))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	if err := proof.VerifyNonMembership(consensusState.GetRoot(), path); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedPacketAckAbsenceVerification, err.Error())
	}

	return nil
}

// VerifyNextSequenceRecv verifies a proof of the next sequence number to be
// received of the specified channel at the specified port.
func (k Keeper) VerifyNextSequenceRecv(
	cs ClientState,
	height uint64,
	prefix commitment.PrefixI,
	proof commitment.ProofI,
	portID,
	channelID string,
	nextSequenceRecv uint64,
	consensusState clientexported.ConsensusState,
) error {
	path, err := commitment.ApplyPrefix(prefix, ibctypes.NextSequenceRecvPath(portID, channelID))
	if err != nil {
		return err
	}

	if err := validateVerificationArgs(cs, height, proof, consensusState); err != nil {
		return err
	}

	bz := sdk.Uint64ToBigEndian(nextSequenceRecv)

	if err := proof.VerifyMembership(consensusState.GetRoot(), path, bz); err != nil {
		return sdkerrors.Wrap(clienttypes.ErrFailedNextSeqRecvVerification, err.Error())
	}

	return nil
}

// validateVerificationArgs perfoms the basic checks on the arguments that are
// shared between the verification functions.
func validateVerificationArgs(
	cs ClientState,
	height uint64,
	proof commitment.ProofI,
	consensusState clientexported.ConsensusState,
) error {
	if cs.LatestHeight < height {
		return sdkerrors.Wrap(
			ibctypes.ErrInvalidHeight,
			fmt.Sprintf("client state (%s) height < proof height (%d < %d)", cs.ID, cs.LatestHeight, height),
		)
	}

	if cs.IsFrozen() && cs.FrozenHeight <= height {
		return clienttypes.ErrClientFrozen
	}

	if proof == nil {
		return sdkerrors.Wrap(commitment.ErrInvalidProof, "proof cannot be empty")
	}

	if consensusState == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state cannot be empty")
	}

	return nil
}
